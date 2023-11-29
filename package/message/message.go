package message

import (
	"collab-net-v2/api"
	pb "collab-net-v2/package/message/proto"
	"context"
	"github.com/pkg/errors"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type MsgCtl struct {
	//ctx    context.Context
	//client pb.MessageClient
	conn *grpc.ClientConn
}

var msg MsgCtl

func GetMsgCtl() *MsgCtl {
	return &msg
}

func (ctl *MsgCtl) Init(address string) (suc bool) {
	var err error
	ctl.conn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("did not connect: %v", err)
	}

	return true
}

func (ctl *MsgCtl) Uninit(address string) {
	defer ctl.conn.Close()
}

func (ctl *MsgCtl) UpdateTaskWrapper(taskId string, status int, extra string) (er error) {
	e := ctl.updateTask(taskId, status, extra)
	if e != nil {
		return errors.Wrap(e, "UpdateTask: ")
	}

	//extra = fmt.Sprintf("Tracking : %s, NewMsg: %s", taskId, extra)
	//e = ctl.updateTask(api.TOPIC_ALL, status, extra)
	//if e != nil {
	//	return errors.Wrap(e, "UpdateTask TOPIC_ALL: ")
	//}

	return nil
}

func (ctl *MsgCtl) updateTask(taskId string, status int, extra string) (e error) {
	client := pb.NewMessageClient(ctl.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if status == api.TASK_STATUS_RUNNING {
		_, err := client.FeedSessionStream(
			ctx,
			&pb.FeedSessionStreamReq{
				SessionId: taskId,
				Timestamp: time.Now().UnixNano() / 1e6,
				Payload:   extra,
			})
		if err != nil {
			logrus.Errorf("FeedSessionStream gRPC err: %v", err) // todo P1 : add extra base64
			logrus.Errorf("FeedSessionStream len(extra): %v", len(extra))
			logrus.Errorf("FeedSessionStream extra: %v", extra)

			return errors.Wrap(err, "UpdateTask: ")
		}
	} else {
		_, err := client.UpdateSessionStatus(
			ctx,
			&pb.UpdateSessionStatusReq{
				SessionId: taskId,
				Timestamp: time.Now().UnixNano() / 1e6,
				EvtType:   int32(status), // model.TASK_STATUS_STARTING,
				Payload:   extra,
			})
		if err != nil {
			logrus.Errorf("UpdateSessionStatus gRPC err: %v", err) // todo P1 : add extra base64
			return errors.Wrap(err, "UpdateTask: ")
		}
	}

	return nil
}

func (ctl *MsgCtl) GetTaskIsHot(taskId string) (isHot int, e error) {
	client := pb.NewMessageClient(ctl.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	resp, err := client.GetSessionStatus(
		ctx,
		&pb.GetSessionStatusReq{
			SessionId: taskId,
		})
	if err != nil {
		logrus.Errorf("GetTaskIsHot gRPC err: %v", err) // todo P1 : add extra base64
		return api.FALSE, errors.Wrap(err, "GetSessionStatus: ")
	}

	if resp.Code != 0 {
		logrus.Errorf("resp.Code != 0: %#v", resp.Code) // todo P1 : add extra base64
		return api.FALSE, errors.Wrap(err, "GetSessionStatus: ")
	}

	if resp.Data == api.TRUE {
		return api.TRUE, nil
	} else if resp.Data == api.FALSE {
		return api.FALSE, nil
	}

	return
}
