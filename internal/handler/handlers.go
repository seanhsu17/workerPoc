package handler

import "github.com/17media/category-lib/taskqueue"

type Handler interface {
	GetTaskMapping() taskqueue.TaskHandlerMap
}
