package common

import "context"

type RunnablePlugin interface {
	Run(context.Context) error
}
