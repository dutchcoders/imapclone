package cmd

import (
	_ "log"

	"go.dutchsec.com/imapclone/cmd/queue"
)

type TaskStack struct {
	*queue.Stack
}

func NewTaskStack() *TaskStack {
	q := queue.NewStack()

	return &TaskStack{
		Stack: q,
	}
}

func (q *TaskStack) Push(v *Task) {
	q.Stack.Push(v)
}

func (q *TaskStack) Pop() *Task {
	v := q.Stack.Pop()
	if v == nil {
		return nil

	}
	return v.(*Task)
}

type Task struct {
	Server   string
	Username string
	Password string

	FolderName string
}
