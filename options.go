package rediswatcher

import (
	"errors"
	"github.com/casbin/casbin/v2"
	rds "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"log"
	"strconv"
	"strings"
)

type WatcherOptions struct {
	Options                rds.Options
	ClusterOptions         rds.ClusterOptions
	SubClient              *rds.Client
	PubClient              *rds.Client
	Channel                string
	IgnoreSelf             bool
	LocalID                string
	OptionalUpdateCallback func(string)
}

func initConfig(option *WatcherOptions) {
	if option.LocalID == "" {
		option.LocalID = uuid.New().String()
	}
	if option.Channel == "" {
		option.Channel = "/casbin"
	}
}

func MakeDefaultUpdateCallback(e casbin.IEnforcer) func(string) {
	return func(msg string) {
		msgStruct := &MSG{}

		err := msgStruct.UnmarshalBinary([]byte(msg))
		if err != nil {
			log.Println(err)
		}

		switch msgStruct.Method {
		case UpdateType_Update:
		case UpdateType_UpdateForSavePolicy:
			err = e.LoadPolicy()
		case UpdateType_UpdateForAddPolicy:
			params := msgStruct.Params[0]

			_, err = e.SelfAddPolicy(msgStruct.Sec, msgStruct.Ptype, params)
		case UpdateType_UpdateForRemovePolicy:
			params := msgStruct.Params[0]

			_, err = e.SelfRemovePolicy(msgStruct.Sec, msgStruct.Ptype, params)
		case UpdateType_UpdateForRemoveFilteredPolicy:
			params := msgStruct.Params[0][0]

			// parse the result of fmt.Sprintf("%d %s", fieldIndex, strings.Join(fieldValues, " "))
			paramsList := strings.Fields(params)
			fieldIndex, _ := strconv.Atoi(paramsList[0])
			fieldValues := paramsList[1:]

			_, err = e.SelfRemoveFilteredPolicy(msgStruct.Sec, msgStruct.Ptype, fieldIndex, fieldValues...)
		case UpdateType_UpdateForAddPolicies:
			params := msgStruct.Params

			_, err = e.SelfAddPolicies(msgStruct.Sec, msgStruct.Ptype, params)
		case UpdateType_UpdateForRemovePolicies:
			params := msgStruct.Params

			_, err = e.SelfRemovePolicies(msgStruct.Sec, msgStruct.Ptype, params)
		default:
			err = errors.New("unknown update type")
		}
		if err != nil {
			log.Println(err)
		}

	}
}
