package live

import (
	"context"
	"net/http"
	"sync"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
)

type StreamManager struct {
	log           log.Logger
	streams       map[string]*Stream
	streamRWMutex *sync.RWMutex
	hub           *hub
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		hub:           newHub(),
		log:           log.New("stream.manager"),
		streams:       make(map[string]*Stream),
		streamRWMutex: &sync.RWMutex{},
	}
}

func (sm *StreamManager) Run(context context.Context) {
	log.Info("初始化流管理器")

	go func() {
		sm.hub.run(context)
		log.Info("停止流管理器")
	}()
}

func (sm *StreamManager) Serve(w http.ResponseWriter, r *http.Request) {
	sm.log.Info("升级到WebSocket")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		sm.log.Error("无法升级到WebSocket的连接", "error", err)
		return
	}

	c := newConnection(ws, sm.hub)
	sm.hub.register <- c

	go c.writePump()
	c.readPump()
}

func (s *StreamManager) GetStreamList() m.StreamList {
	list := make(m.StreamList, 0)

	for _, stream := range s.streams {
		list = append(list, &m.StreamInfo{
			Name: stream.name,
		})
	}

	return list
}

func (s *StreamManager) Push(packet *m.StreamPacket) {
	stream, exist := s.streams[packet.Stream]

	if !exist {
		s.log.Info("Creating metric stream", "name", packet.Stream)
		stream = NewStream(packet.Stream)
		s.streams[stream.name] = stream
	}

	stream.Push(packet)
}

type Stream struct {
	subscribers []*connection
	name        string
}

func NewStream(name string) *Stream {
	return &Stream{
		subscribers: make([]*connection, 0),
		name:        name,
	}
}

func (s *Stream) Push(packet *m.StreamPacket) {

	messageBytes, _ := simplejson.NewFromAny(packet).Encode()

	for _, sub := range s.subscribers {
		// check if channel is open
		// if _, ok := h.connections[sub]; !ok {
		// 	delete(s.subscribers, sub)
		// 	continue
		// }

		sub.send <- messageBytes
	}
}
