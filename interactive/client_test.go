package main

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestGRPCClient(t *testing.T) {
	conn, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := interv1.NewInteractiveServiceClient(conn)
	resp, err := client.Get(context.Background(), &interv1.GetRequest{
		Biz:   "test",
		BizId: 7,
		Uid:   777,
	})
	require.NoError(t, err)
	resp2, err := client.GetByIds(context.Background(), &interv1.GetByIdsRequest{
		Biz:    "test",
		BizIds: []int64{1, 3, 5, 7, 9},
	})
	require.NoError(t, err)
	t.Log(resp.Inter)
	t.Log(resp2.Inters)
}

// status Mysql
//TRUNCATE table interactives;
//
//INSERT INTO `interactives`(`biz_id`,`biz`,`read_cnt`,`collect_cnt`,`like_cnt`,`create_time`,`update_time`)
//VALUES
//    (1,"test",6142,1089,2454,1701501220644,1701501220644),
//    (2,"test",7542,2222,3321,1701501230644,1701511220644),
//    (3,"test",9132,1000,4121,1701501240644,1701521220644),
//    (4,"test",9764,9000,1444,1701501250644,1701531220644),
//    (5,"test",3341,3000,2411,1701501260644,1701541220644),
//    (6,"test",6632,6000,2754,1701501270644,1701551220644),
//    (7,"test",6552,5000,1331,1701501280644,1701561220644),
//    (8,"test",1131,1000,1000,1701501290644,1701571220644),
//    (9,"test",1111,1000,1020,1701501300644,1701581220644),
//    (10,"test",9999,9999,9999,1701509999644,1701589999644);
//
//SELECT * FROM `interactives`;
