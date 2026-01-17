package session

type holder struct {
	session *Session
}

func (h *holder) get() *Session {
	return h.session
}

func (h *holder) set(sess *Session) {
	h.session = sess
}
