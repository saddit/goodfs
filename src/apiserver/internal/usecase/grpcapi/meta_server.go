package grpcapi

import (
	"apiserver/internal/entity"
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
	if err = ResolveErr(err); err != nil {
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
	if err = ResolveErr(err); err != nil {
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
	if err = ResolveErr(err); err != nil {
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
	resp, err := cli.GetPeers(ctx, &pb.EmptyReq{})
	if err = ResolveErr(err); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
