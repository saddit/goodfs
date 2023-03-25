package grpcapi

import (
	"apiserver/internal/entity"
	"common/proto"
	"common/proto/msg"
	"common/proto/pb"
	"common/util"
	"context"
)

func GetMetadata(ip, id string, withExtra bool) (*entity.Metadata, error) {
	defer perform(false)()
	conn, err := getConn(ip)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli := pb.NewMetadataApiClient(conn)
	resp, err := cli.GetMetadata(ctx, &pb.MetaReq{Id: id, WithExtra: withExtra})
	if err = proto.ResolveErr(err); err != nil {
		return nil, err
	}
	var m msg.Metadata
	if err = util.DecodeMsgp(&m, resp.Data); err != nil {
		return nil, err
	}
	res := &entity.Metadata{
		Name:       m.Name,
		Bucket:     m.Bucket,
		CreateTime: m.CreateTime,
		UpdateTime: m.UpdateTime,
	}
	if withExtra && m.Extra != nil {
		res.Extra = entity.Extra{
			Total:        m.Extra.Total,
			FirstVersion: m.Extra.FirstVersion,
			LastVersion:  m.Extra.LastVersion,
		}
	}
	return res, nil
}

func GetVersion(ip, id string, verNum int32) (*entity.Version, error) {
	defer perform(false)()
	conn, err := getConn(ip)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli := pb.NewMetadataApiClient(conn)
	resp, err := cli.GetVersion(ctx, &pb.MetaReq{Id: id, Version: verNum})
	if err = proto.ResolveErr(err); err != nil {
		return nil, err
	}
	var v msg.Version
	if err = util.DecodeMsgp(&v, resp.Data); err != nil {
		return nil, err
	}
	return &entity.Version{
		Compress:      v.Compress,
		Hash:          v.Hash,
		StoreStrategy: entity.ObjectStrategy(v.StoreStrategy),
		Sequence:      int32(v.Sequence),
		Size:          v.Size,
		Ts:            v.Ts,
		DataShards:    int(v.DataShards),
		ParityShards:  int(v.ParityShards),
		ShardSize:     int(v.ShardSize),
		Locate:        v.Locate,
	}, nil
}

func GetBucket(ip, name string) (*entity.Bucket, error) {
	defer perform(false)()
	conn, err := getConn(ip)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli := pb.NewMetadataApiClient(conn)
	resp, err := cli.GetBucket(ctx, &pb.MetaReq{Id: name})
	if err = proto.ResolveErr(err); err != nil {
		return nil, err
	}
	var b msg.Bucket
	if err = util.DecodeMsgp(&b, resp.Data); err != nil {
		return nil, err
	}
	return &entity.Bucket{
		Versioning:     b.Versioning,
		Readonly:       b.Readonly,
		Compress:       b.Compress,
		StoreStrategy:  entity.ObjectStrategy(b.StoreStrategy),
		DataShards:     int(b.DataShards),
		ParityShards:   int(b.ParityShards),
		VersionRemains: int(b.VersionRemains),
		CreateTime:     b.CreateTime,
		UpdateTime:     b.UpdateTime,
		Name:           b.Name,
		Policies:       b.Policies,
	}, nil
}

func GetPeers(ip string) ([]string, error) {
	defer perform(false)()
	conn, err := getConn(ip)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli := pb.NewMetadataApiClient(conn)
	resp, err := cli.GetPeers(ctx, new(pb.Empty))
	if err = proto.ResolveErr(err); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func UpdateVersion(ip, id string, body *entity.Version) error {
	defer perform(true)()
	conn, err := getConn(ip)
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(&msg.Version{
		Compress:      body.Compress,
		StoreStrategy: int8(body.StoreStrategy),
		DataShards:    int32(body.DataShards),
		ParityShards:  int32(body.ParityShards),
		ShardSize:     int64(body.ShardSize),
		Sequence:      uint64(body.Sequence),
		Size:          body.Size,
		Hash:          body.Hash,
		Locate:        body.Locate,
	})
	_, err = pb.NewMetadataApiClient(conn).UpdateVersion(context.Background(), &pb.Metadata{
		Id:      id,
		Msgpack: bt,
	})
	return proto.ResolveErr(err)
}

func SaveVersion(ip, id string, body *entity.Version) (int32, error) {
	defer perform(true)()
	conn, err := getConn(ip)
	if err != nil {
		return 0, err
	}
	bt, err := util.EncodeMsgp(&msg.Version{
		Compress:      body.Compress,
		StoreStrategy: int8(body.StoreStrategy),
		DataShards:    int32(body.DataShards),
		ParityShards:  int32(body.ParityShards),
		ShardSize:     int64(body.ShardSize),
		Size:          body.Size,
		Hash:          body.Hash,
		Locate:        body.Locate,
	})
	res, err := pb.NewMetadataApiClient(conn).SaveVersion(context.Background(), &pb.Metadata{
		Id:      id,
		Version: body.Sequence,
		Msgpack: bt,
	})
	if err != nil {
		return 0, proto.ResolveErr(err)
	}
	return res.Data, nil
}

func SaveMetadata(ip, id string, body *entity.Metadata) error {
	defer perform(true)()
	conn, err := getConn(ip)
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(&msg.Metadata{
		Name:   body.Name,
		Bucket: body.Bucket,
	})
	_, err = pb.NewMetadataApiClient(conn).SaveMetadata(context.Background(), &pb.Metadata{
		Id:      id,
		Msgpack: bt,
	})
	return proto.ResolveErr(err)
}

func SaveBucket(ip string, body *entity.Bucket) error {
	defer perform(true)()
	conn, err := getConn(ip)
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(&msg.Bucket{
		Versioning:     body.Versioning,
		Readonly:       body.Readonly,
		Compress:       body.Compress,
		StoreStrategy:  int8(body.StoreStrategy),
		DataShards:     int32(body.DataShards),
		ParityShards:   int32(body.ParityShards),
		VersionRemains: int32(body.VersionRemains),
		Name:           body.Name,
		Policies:       body.Policies,
	})
	_, err = pb.NewMetadataApiClient(conn).SaveBucket(context.Background(), &pb.Metadata{
		Id:      body.Name,
		Msgpack: bt,
	})
	return proto.ResolveErr(err)
}

func ListVersion(ip, id string, page, pageSize int) (arr []*msg.Version, total int64, err error) {
	defer perform(false)()
	conn, err := getConn(ip)
	if err != nil {
		return
	}
	resp, err := pb.NewMetadataApiClient(conn).ListVersion(context.Background(), &pb.MetaReq{
		Id: id,
		Page: &pb.Pageable{
			Page:     int32(page),
			PageSize: int32(pageSize),
		},
	})
	if err != nil {
		err = proto.ResolveErr(err)
		return
	}
	total = resp.Total
	arr, err = util.DecodeArrayMsgp(resp.Data, func() *msg.Version { return new(msg.Version) })
	return
}

func RemoveVersion(ip, id string, version int32) error {
	defer perform(true)()
	conn, err := getConn(ip)
	if err != nil {
		return err
	}
	_, err = pb.NewMetadataApiClient(conn).RemoveVersion(context.Background(), &pb.MetaReq{Id: id, Version: version})
	return err
}
