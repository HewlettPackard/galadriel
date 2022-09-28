package admin

import (
	"net"

	"github.com/sirupsen/logrus"
)

const (
	CreateMemberPath       = "/member"
	CreateRelationshipPath = "/relationship"
	GenerateTokenPath      = "/token"
)

type Config struct {
	LocalAddress net.Addr
	Logger       logrus.FieldLogger
}
