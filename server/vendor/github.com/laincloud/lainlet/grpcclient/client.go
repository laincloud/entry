package grpcclient

import (
	"context"
	"fmt"
	"time"

	"sync"

	pb "github.com/laincloud/lainlet/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	RPC_TIMEOUT = 10
)

type Config struct {
	Addr               string
	CertFile           string
	ServerNameOverride string
	Timeout            int
}

type Client struct {
	addr    string
	cfg     *Config
	timeout time.Duration

	lainletClient pb.LainletClient

	appnameClient    pb.AppnameClient
	appsClient       pb.AppsClient
	backupctlClient  pb.BackupctlClient
	configClient     pb.ConfigClient
	containersClient pb.ContainersClient
	coreinfoClient   pb.CoreinfoClient
	dependsClient    pb.DependsClient
	localspecClient  pb.LocalspecClient
	nodesClient      pb.NodesClient
	podgroupClient   pb.PodgroupClient
	proxyClient      pb.ProxyClient

	rebellionLocalprocsClient     pb.RebellionLocalprocsClient
	streamrouterPortsClient       pb.StreamrouterPortsClient
	streamrouterStreamprocsClient pb.StreamrouterStreamprocsClient
	webrouterWebprocsClient       pb.WebrouterWebprocsClient
}

// create a new client, addr is lainlet address such as "127.0.0.1:9001"
func New(cfg *Config) (*Client, error) {
	var opts []grpc.DialOption
	if cfg.CertFile != "" {
		creds, err := credentials.NewClientTLSFromFile(cfg.CertFile, cfg.ServerNameOverride)
		if err != nil {
			return nil, fmt.Errorf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		addr:          cfg.Addr,
		cfg:           cfg,
		lainletClient: pb.NewLainletClient(conn),
		appnameClient: pb.NewAppnameClient(conn),

		appsClient:                    pb.NewAppsClient(conn),
		backupctlClient:               pb.NewBackupctlClient(conn),
		configClient:                  pb.NewConfigClient(conn),
		containersClient:              pb.NewContainersClient(conn),
		coreinfoClient:                pb.NewCoreinfoClient(conn),
		dependsClient:                 pb.NewDependsClient(conn),
		localspecClient:               pb.NewLocalspecClient(conn),
		nodesClient:                   pb.NewNodesClient(conn),
		podgroupClient:                pb.NewPodgroupClient(conn),
		proxyClient:                   pb.NewProxyClient(conn),
		rebellionLocalprocsClient:     pb.NewRebellionLocalprocsClient(conn),
		streamrouterPortsClient:       pb.NewStreamrouterPortsClient(conn),
		streamrouterStreamprocsClient: pb.NewStreamrouterStreamprocsClient(conn),
		webrouterWebprocsClient:       pb.NewWebrouterWebprocsClient(conn),
	}

	if cfg.Timeout <= 0 {
		cli.timeout = RPC_TIMEOUT * time.Second
	} else {
		cli.timeout = time.Duration(cfg.Timeout) * time.Second
	}

	return cli, nil
}

func (cli *Client) Version() (*pb.VersionReply, error) {
	req := &pb.EmptyRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.lainletClient.Version(ctx, req)
	return rpl, err
}

func (cli *Client) Status() (*pb.StatusReply, error) {
	req := &pb.EmptyRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.lainletClient.Status(ctx, req)
	return rpl, err
}

// Appname only has Get rpc
func (cli *Client) AppnameGet(ip string) (*pb.AppnameReply, error) {
	req := &pb.AppnameRequest{Ip: ip}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.appnameClient.Get(ctx, req)
	return rpl, err
}

// Localspec only has Get rpc
func (cli *Client) LocalspecGet(nodeip string) (*pb.LocalspecReply, error) {
	req := &pb.LocalspecRequest{Nodeip: nodeip}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.localspecClient.Get(ctx, req)
	return rpl, err
}

//go:generate python ../tools/gen/cli.py Apps
//go:generate python ../tools/gen/cli.py Backupctl appname
//go:generate python ../tools/gen/cli.py Config target
//go:generate python ../tools/gen/cli.py Containers nodename
//go:generate python ../tools/gen/cli.py Coreinfo appname
//go:generate python ../tools/gen/cli.py Depends target
//go:generate python ../tools/gen/cli.py Nodes name
//go:generate python ../tools/gen/cli.py Podgroup appname
//go:generate python ../tools/gen/cli.py Proxy appname
//go:generate python ../tools/gen/cli.py RebellionLocalprocs appname
//go:generate python ../tools/gen/cli.py StreamrouterPorts appname
//go:generate python ../tools/gen/cli.py StreamrouterStreamprocs appname
//go:generate python ../tools/gen/cli.py WebrouterWebprocs appname

//CODE GENERATION Apps START
func (cli *Client) AppsGet() (*pb.AppsReply, error) {
	req := &pb.AppsRequest{}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.appsClient.Get(ctx, req)
	return rpl, err
}

type AppsWatcher struct {
	err    error
	ch     chan *pb.AppsReply
	sync.RWMutex
}

func (wch *AppsWatcher) Next() (*pb.AppsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) AppsWatch() (*AppsWatcher, error) {
	req := &pb.AppsRequest{}
	stream, err := cli.appsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &AppsWatcher{
		ch: make(chan *pb.AppsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Apps END
//CODE GENERATION Backupctl START
func (cli *Client) BackupctlGet(appname string) (*pb.BackupctlReply, error) {
	req := &pb.BackupctlRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.backupctlClient.Get(ctx, req)
	return rpl, err
}

type BackupctlWatcher struct {
	err    error
	ch     chan *pb.BackupctlReply
	sync.RWMutex
}

func (wch *BackupctlWatcher) Next() (*pb.BackupctlReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) BackupctlWatch(appname string) (*BackupctlWatcher, error) {
	req := &pb.BackupctlRequest{Appname: appname}
	stream, err := cli.backupctlClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &BackupctlWatcher{
		ch: make(chan *pb.BackupctlReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Backupctl END
//CODE GENERATION Config START
func (cli *Client) ConfigGet(target string) (*pb.ConfigReply, error) {
	req := &pb.ConfigRequest{Target: target}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.configClient.Get(ctx, req)
	return rpl, err
}

type ConfigWatcher struct {
	err    error
	ch     chan *pb.ConfigReply
	sync.RWMutex
}

func (wch *ConfigWatcher) Next() (*pb.ConfigReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) ConfigWatch(target string) (*ConfigWatcher, error) {
	req := &pb.ConfigRequest{Target: target}
	stream, err := cli.configClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &ConfigWatcher{
		ch: make(chan *pb.ConfigReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Config END
//CODE GENERATION Containers START
func (cli *Client) ContainersGet(nodename string) (*pb.ContainersReply, error) {
	req := &pb.ContainersRequest{Nodename: nodename}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.containersClient.Get(ctx, req)
	return rpl, err
}

type ContainersWatcher struct {
	err    error
	ch     chan *pb.ContainersReply
	sync.RWMutex
}

func (wch *ContainersWatcher) Next() (*pb.ContainersReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) ContainersWatch(nodename string) (*ContainersWatcher, error) {
	req := &pb.ContainersRequest{Nodename: nodename}
	stream, err := cli.containersClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &ContainersWatcher{
		ch: make(chan *pb.ContainersReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Containers END
//CODE GENERATION Coreinfo START
func (cli *Client) CoreinfoGet(appname string) (*pb.CoreinfoReply, error) {
	req := &pb.CoreinfoRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.coreinfoClient.Get(ctx, req)
	return rpl, err
}

type CoreinfoWatcher struct {
	err    error
	ch     chan *pb.CoreinfoReply
	sync.RWMutex
}

func (wch *CoreinfoWatcher) Next() (*pb.CoreinfoReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) CoreinfoWatch(appname string) (*CoreinfoWatcher, error) {
	req := &pb.CoreinfoRequest{Appname: appname}
	stream, err := cli.coreinfoClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &CoreinfoWatcher{
		ch: make(chan *pb.CoreinfoReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Coreinfo END
//CODE GENERATION Nodes START
func (cli *Client) NodesGet(name string) (*pb.NodesReply, error) {
	req := &pb.NodesRequest{Name: name}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.nodesClient.Get(ctx, req)
	return rpl, err
}

type NodesWatcher struct {
	err    error
	ch     chan *pb.NodesReply
	sync.RWMutex
}

func (wch *NodesWatcher) Next() (*pb.NodesReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) NodesWatch(name string) (*NodesWatcher, error) {
	req := &pb.NodesRequest{Name: name}
	stream, err := cli.nodesClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &NodesWatcher{
		ch: make(chan *pb.NodesReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Nodes END
//CODE GENERATION Podgroup START
func (cli *Client) PodgroupGet(appname string) (*pb.PodgroupReply, error) {
	req := &pb.PodgroupRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.podgroupClient.Get(ctx, req)
	return rpl, err
}

type PodgroupWatcher struct {
	err    error
	ch     chan *pb.PodgroupReply
	sync.RWMutex
}

func (wch *PodgroupWatcher) Next() (*pb.PodgroupReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) PodgroupWatch(appname string) (*PodgroupWatcher, error) {
	req := &pb.PodgroupRequest{Appname: appname}
	stream, err := cli.podgroupClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &PodgroupWatcher{
		ch: make(chan *pb.PodgroupReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Podgroup END
//CODE GENERATION Proxy START
func (cli *Client) ProxyGet(appname string) (*pb.ProxyReply, error) {
	req := &pb.ProxyRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.proxyClient.Get(ctx, req)
	return rpl, err
}

type ProxyWatcher struct {
	err    error
	ch     chan *pb.ProxyReply
	sync.RWMutex
}

func (wch *ProxyWatcher) Next() (*pb.ProxyReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) ProxyWatch(appname string) (*ProxyWatcher, error) {
	req := &pb.ProxyRequest{Appname: appname}
	stream, err := cli.proxyClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &ProxyWatcher{
		ch: make(chan *pb.ProxyReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Proxy END
//CODE GENERATION RebellionLocalprocs START
func (cli *Client) RebellionLocalprocsGet(appname string) (*pb.RebellionLocalprocsReply, error) {
	req := &pb.RebellionLocalprocsRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.rebellionLocalprocsClient.Get(ctx, req)
	return rpl, err
}

type RebellionLocalprocsWatcher struct {
	err    error
	ch     chan *pb.RebellionLocalprocsReply
	sync.RWMutex
}

func (wch *RebellionLocalprocsWatcher) Next() (*pb.RebellionLocalprocsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) RebellionLocalprocsWatch(appname string) (*RebellionLocalprocsWatcher, error) {
	req := &pb.RebellionLocalprocsRequest{Appname: appname}
	stream, err := cli.rebellionLocalprocsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &RebellionLocalprocsWatcher{
		ch: make(chan *pb.RebellionLocalprocsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION RebellionLocalprocs END
//CODE GENERATION StreamrouterPorts START
func (cli *Client) StreamrouterPortsGet(appname string) (*pb.StreamrouterPortsReply, error) {
	req := &pb.StreamrouterPortsRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.streamrouterPortsClient.Get(ctx, req)
	return rpl, err
}

type StreamrouterPortsWatcher struct {
	err    error
	ch     chan *pb.StreamrouterPortsReply
	sync.RWMutex
}

func (wch *StreamrouterPortsWatcher) Next() (*pb.StreamrouterPortsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) StreamrouterPortsWatch(appname string) (*StreamrouterPortsWatcher, error) {
	req := &pb.StreamrouterPortsRequest{Appname: appname}
	stream, err := cli.streamrouterPortsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &StreamrouterPortsWatcher{
		ch: make(chan *pb.StreamrouterPortsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION StreamrouterPorts END
//CODE GENERATION StreamrouterStreamprocs START
func (cli *Client) StreamrouterStreamprocsGet(appname string) (*pb.StreamrouterStreamprocsReply, error) {
	req := &pb.StreamrouterStreamprocsRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.streamrouterStreamprocsClient.Get(ctx, req)
	return rpl, err
}

type StreamrouterStreamprocsWatcher struct {
	err    error
	ch     chan *pb.StreamrouterStreamprocsReply
	sync.RWMutex
}

func (wch *StreamrouterStreamprocsWatcher) Next() (*pb.StreamrouterStreamprocsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) StreamrouterStreamprocsWatch(appname string) (*StreamrouterStreamprocsWatcher, error) {
	req := &pb.StreamrouterStreamprocsRequest{Appname: appname}
	stream, err := cli.streamrouterStreamprocsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &StreamrouterStreamprocsWatcher{
		ch: make(chan *pb.StreamrouterStreamprocsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION StreamrouterStreamprocs END
//CODE GENERATION WebrouterWebprocs START
func (cli *Client) WebrouterWebprocsGet(appname string) (*pb.WebrouterWebprocsReply, error) {
	req := &pb.WebrouterWebprocsRequest{Appname: appname}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.webrouterWebprocsClient.Get(ctx, req)
	return rpl, err
}

type WebrouterWebprocsWatcher struct {
	err    error
	ch     chan *pb.WebrouterWebprocsReply
	sync.RWMutex
}

func (wch *WebrouterWebprocsWatcher) Next() (*pb.WebrouterWebprocsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) WebrouterWebprocsWatch(appname string) (*WebrouterWebprocsWatcher, error) {
	req := &pb.WebrouterWebprocsRequest{Appname: appname}
	stream, err := cli.webrouterWebprocsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &WebrouterWebprocsWatcher{
		ch: make(chan *pb.WebrouterWebprocsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION WebrouterWebprocs END

//CODE GENERATION Depends START
func (cli *Client) DependsGet(target string) (*pb.DependsReply, error) {
	req := &pb.DependsRequest{Target: target}
	ctx, cancel := context.WithTimeout(context.Background(), cli.timeout)
	defer cancel()
	rpl, err := cli.dependsClient.Get(ctx, req)
	return rpl, err
}

type DependsWatcher struct {
	err    error
	ch     chan *pb.DependsReply
	sync.RWMutex
}

func (wch *DependsWatcher) Next() (*pb.DependsReply, error) {
	rpl := <- wch.ch
	wch.RLock()
	defer wch.RUnlock()
	err := wch.err
	return rpl, err
}
func (cli *Client) DependsWatch(target string) (*DependsWatcher, error) {
	req := &pb.DependsRequest{Target: target}
	stream, err := cli.dependsClient.Watch(context.Background(), req)
	if err != nil {
		return nil, err
	}
	wch := &DependsWatcher{
		ch: make(chan *pb.DependsReply),
	}
	go func() {
		defer close(wch.ch)
		for {
			in, err := stream.Recv()
			if err != nil {
				wch.Lock()
				wch.err = err
				wch.Unlock()
				return
			}
			wch.ch <- in
		}
	}()
	return wch, nil
}
//CODE GENERATION Depends END
