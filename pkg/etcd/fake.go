package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type fakeClient struct {
	memberListF   func(context context.Context) (*clientv3.MemberListResponse, error)
	memberAddF    func(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error)
	memberRemoveF func(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error)
}

func (f *fakeClient) MemberList(context context.Context) (*clientv3.MemberListResponse, error) {
	return f.memberListF(context)
}

func (f *fakeClient) MemberAdd(context context.Context, peerURLs []string) (*clientv3.MemberAddResponse, error) {
	return f.memberAddF(context, peerURLs)
}

func (f *fakeClient) MemberRemove(context context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	return f.memberRemoveF(context, id)
}

func (f *fakeClient) Close() error {
	return nil
}
