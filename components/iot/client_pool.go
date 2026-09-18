/**
 * 已发布流程的 MQTT 发布客户端池（由 flowgo-server 注册）。
 * mqttOut 在「复用收节点」或「常驻」模式下从此池取连接；临时模式自行短连。
 */
package iot

import (
	"fmt"
	"sync"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// PublisherPool 按 flowID + mqttOut 节点 id 提供已连接的发布客户端。
type PublisherPool interface {
	// Publisher 返回托管客户端；无托管（临时模式或未上线）时 ok=false。
	Publisher(flowID, outNodeID string) (client paho.Client, ok bool)
}

var (
	poolMu sync.RWMutex
	pool   PublisherPool
)

// SetPublisherPool 由服务端在启动时注册；传 nil 清空。
func SetPublisherPool(p PublisherPool) {
	poolMu.Lock()
	defer poolMu.Unlock()
	pool = p
}

// LookupPublisher 查找托管发布客户端。
func LookupPublisher(flowID, outNodeID string) (paho.Client, bool) {
	poolMu.RLock()
	p := pool
	poolMu.RUnlock()
	if p == nil || flowID == "" || outNodeID == "" {
		return nil, false
	}
	return p.Publisher(flowID, outNodeID)
}

// ErrNoManagedPublisher 需要托管客户端但尚未上线/未建立。
var ErrNoManagedPublisher = fmt.Errorf("mqttOut: managed client unavailable (publish the flow first)")
