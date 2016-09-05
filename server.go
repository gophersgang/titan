package titan

import (
	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
	"github.com/neptulon/neptulon/middleware/jwt"
	"github.com/titan-x/titan/data"
	"github.com/titan-x/titan/data/inmem"
)

// Server wraps a listener instance and registers default connection and message handlers with the listener.
type Server struct {
	// neptulon framework components
	server     *neptulon.Server
	pubRoutes  *middleware.Router
	privRoutes *middleware.Router

	// titan server components
	db    data.DB
	queue Queue
}

// NewServer creates a new server.
func NewServer(addr string) (*Server, error) {
	if (Conf == Config{}) {
		InitConf("")
	}

	s := Server{
		server: neptulon.NewServer(addr),
		db:     inmem.NewDB(),
		queue:  NewQueue(),
	}

	s.server.MiddlewareFunc(middleware.Logger)
	s.pubRoutes = middleware.NewRouter()
	s.server.Middleware(s.pubRoutes)
	initPubRoutes(s.pubRoutes, s.db, Conf.App.JWTPass())

	//all communication below this point is authenticated
	s.server.MiddlewareFunc(jwt.HMAC(Conf.App.JWTPass()))
	s.server.Middleware(&s.queue)
	s.privRoutes = middleware.NewRouter()
	s.server.Middleware(s.privRoutes)
	initPrivRoutes(s.privRoutes, &s.queue)
	// r.Middleware(NotFoundHandler()) - 404-like handler

	// todo: research a better way to handle inner-circular dependencies so remove these lines back into Server contructor
	// (maybe via dereferencing: http://openmymind.net/Things-I-Wish-Someone-Had-Told-Me-About-Go/, but then initializers
	// actually using the pointer values would have to be lazy!)
	s.queue.SetServer(s.server)

	s.server.DisconnHandler(func(c *neptulon.Conn) {
		// only handle this event for previously authenticated
		if id, ok := c.Session.GetOk("userid"); ok {
			s.queue.RemoveConn(id.(string))
		}
	})

	return &s, nil
}

// SetDB sets the database to be used by the server. If not supplied, in-memory database implementation is used.
func (s *Server) SetDB(db data.DB) error {
	s.db = db
	return nil
}

// ListenAndServe starts the Titan server. This function blocks until server is closed.
func (s *Server) ListenAndServe() error {
	if err := s.db.Seed(false, Conf.App.JWTPass()); err != nil {
		return err
	}

	return s.server.ListenAndServe()
}

// Close the server and all of the active connections, discarding any read/writes that is going on currently.
// This is not a problem as we always require an ACK but it will also mean that message deliveries will be at-least-once; to-and-from the server.
func (s *Server) Close() error {
	return s.server.Close()
}
