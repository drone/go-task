package heartbeat

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/icrowley/fake"
	"github.com/drone/go-task/delegateshell/client"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// Time period between sending heartbeats to the server
	hearbeatInterval  = 10 * time.Second
	heartbeatTimeout  = 15 * time.Second
	taskEventsTimeout = 30 * time.Second
)

type FilterFn func(*client.TaskEvent) bool

type KeepAlive struct {
	AccountID     string
	AccountSecret string
	Name          string   // name of the runner
	Tags          []string // list of tags that the runner accepts
	Client        client.Client
	Filter        FilterFn
	// The Harness manager allows two task acquire calls with the same delegate ID to go through (by design).
	// We need to make sure two different threads do not acquire the same task.
	// This map makes sure Acquire() is called only once per task ID. The mapping is removed once the status
	// for the task has been sent.
	m sync.Map
}

type DelegateInfo struct {
	Host string
	IP   string
	ID   string
	Name string
}

func New(accountID, accountSecret, name string, tags []string, c client.Client) *KeepAlive {
	return &KeepAlive{
		AccountID:     accountID,
		AccountSecret: accountSecret,
		Tags:          tags,
		Name:          name,
		Client:        c,
		m:             sync.Map{},
	}
}

func (p *KeepAlive) SetFilter(filter FilterFn) {
	p.Filter = filter
}

// Register registers the runner with the server. The server generates a delegate ID
// which is returned to the client.
func (p *KeepAlive) Register(ctx context.Context) (*DelegateInfo, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "could not get host name")
	}
	host = "dlite-" + strings.ReplaceAll(host, " ", "-")
	ip := getOutboundIP()
	id, err := p.register(ctx, hearbeatInterval, ip, host)
	if err != nil {
		logrus.WithField("ip", ip).WithField("host", host).WithError(err).Error("could not register runner")
		return nil, err
	}
	return &DelegateInfo{
		ID:   id,
		Host: host,
		IP:   ip,
		Name: p.Name,
	}, nil
}

// Register registers the runner and runs a background thread which keeps pinging the server
// at a period of interval. It returns the delegate ID.
func (p *KeepAlive) register(ctx context.Context, interval time.Duration, ip, host string) (string, error) {
	req := &client.RegisterRequest{
		AccountID:     p.AccountID,
		RunnerName:    p.Name,
		LastHeartbeat: time.Now().UnixMilli(),
		//Token:              p.AccountSecret,
		NG:       true,
		Type:     "DOCKER",
		Polling:  true,
		HostName: host,
		IP:       ip,
		// SupportedTaskTypes: p.Router.Routes(),  // Ignore this because for new Runner tasks, this SupportedTaskTypes feature doesn't apply
		Tags:              p.Tags,
		HeartbeatAsObject: true,
	}
	resp, err := p.Client.Register(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "could not register the runner")
	}
	req.ID = resp.Resource.DelegateID
	logrus.WithField("id", req.ID).WithField("host", req.HostName).
		WithField("ip", req.IP).Info("registered delegate successfully")
	p.heartbeat(ctx, req, interval)
	return resp.Resource.DelegateID, nil
}

// heartbeat starts a periodic thread in the background which continually pings the server
func (p *KeepAlive) heartbeat(ctx context.Context, req *client.RegisterRequest, interval time.Duration) {
	go func() {
		msgDelayTimer := time.NewTimer(interval)
		defer msgDelayTimer.Stop()
		for {
			msgDelayTimer.Reset(interval)
			select {
			case <-ctx.Done():
				logrus.Error("context canceled")
				return
			case <-msgDelayTimer.C:
				req.LastHeartbeat = time.Now().UnixMilli()
				heartbeatCtx, cancelFn := context.WithTimeout(ctx, heartbeatTimeout)
				err := p.Client.Heartbeat(heartbeatCtx, req)
				if err != nil {
					logrus.WithError(err).Errorf("could not send heartbeat")
				}
				cancelFn()
			}
		}
	}()
}

// Get preferred outbound ip of this machine. It returns a fake IP in case of errors.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logrus.WithError(err).Error("could not figure out an IP, using a randomly generated IP")
		return "fake-" + fake.IPv4()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
