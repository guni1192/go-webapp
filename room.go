package main

type room struct {
  // 他のクライアントに転送するためのメッセージを保持する
  forward chan []byte
  // チャットルームに参加しようとしているクライアントのチャネル
  join chan *client
  // チャットルームから退出しようとしているクライアントのチャネル
  leave chan *client
  // 在室しているすべてのクライアント
  clients map[*client]bool
}

func newRoom() *room {
  return &room {
    foward: make(chan []byte),
  }
}

func (r *room) run() {
  for {
    select {
    case client := <-r.join:
      r.clients[client] = true
    case client := <-r.leave:
      delete(r.clients, r.forward)
      close(client.send)
    case msg := <-r.forward:
      for client := range r.clients {
        select {
        case client.send <- msg:
          //  メッセージ送信
        default:
          delete(r.clients, client)
          close(client.send)
        }
      }
    }
  }
}

const (
  socketBufferSize = 1024
  messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
  ReadBufferSize: socketBufferSize,
  WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  socket, err := upgrader.Upgrader(w, req, nil)
  if err != nil {
    log.Fatal("ServeHTTP: ", err)
    return
  }

  client := &client {
    socket: socket,
    send: make(chan []byte, messageBufferSize),
    room: r,
  }

  r.join <-client
  defer func() { r.leave <- client }()
  go client.write()
  client.read()
}

