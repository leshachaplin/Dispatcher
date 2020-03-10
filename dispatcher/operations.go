package dispatcher

type Operation interface {
}

type ReadRoleOperation struct {
	Role string
}

type WriteRole struct {
}

type ReadMessageOperation struct {
}

type WriteMessageOperation struct {
}

type CancelWrite struct {
}

type CancelRead struct {
}
