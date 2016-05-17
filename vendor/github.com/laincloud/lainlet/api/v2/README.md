
v2版的API的Data格式统一成了json, 各个API的返回结果的Data的源数据如下, 可使用对应数据结构的`xxx.Decode([]byte)`函数进行解析

使用这些数据结构需要:

```golang
import "github.com/laincloud/lainlet/api/v2"
```

## /v2/configwatcher

```golang
type GeneralConfig struct {
	Data map[string]string // data type return by configwatcher
}
```

## /v2/coreinfowatcher

```golang
import lainEngine "github.com/laincloud/deployd/engine"

type CoreInfo lainEngine.PodGroupCoreInfo // 这个格式需要参考deployd的代码

type GeneralCoreInfo struct {
	Data map[string]coreinfo.CoreInfo // 真正的Data
}
```

## /v2/localspecquery

```golang
type LocalSpec struct {
	Data    []string // 真正的Data
}
```

## /v2/procwatcher

```golang
type Pod struct {
	InstanceNo int
	IP         string
	Port       int
	ProcName   string
}

type PodGroup struct {
	Pods []Pod `json:"proc"`
}

type GeneralPodGroup struct {
	Data []PodGroup // 真正的Data
}
```

## /v2/proxywatcher

```golang
type Container struct {
	ContainerIp   string `json:"container_ip"`
	ContainerPort int    `json:"container_port"`
}

type ProcInfo struct {
	Containers []Container `json:"containers"`
}

type ProxyData struct {    
	Data map[string]ProcInfo // 真正的data
}

```

## /v2/depends

```golang
type ContainerInfo struct {
	NodeIP string
	IP     string
	Port   int
}

type DependsItem struct {
	Annotation string
	Containers []ContainerInfo
}

type Depends struct {
	Data map[string]map[string]map[string]DependsItem // 真正的data
}


```
