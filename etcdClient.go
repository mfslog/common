package common

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

var etcdIns *etcdClient

var onceEtcd sync.Once

//单例返回 DBCconnect Instance
func GetEtcdConIns() *etcdClient {
	onceEtcd.Do(func() {
		etcdIns = &etcdClient{}
	})
	return etcdIns
}

type etcdClient struct {
	client  *clientv3.Client
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func (etcd *etcdClient) Init(etcdEndPoints []string) (error){
	var err error
	//连接etcd
	err = etcd.connectEtcd(etcdEndPoints)
	if err != nil {
		return err
	}
	etcd.session, err = concurrency.NewSession(etcd.client)
	return  err
}

//连接etcd 服务
//endPoints 为etcd endpoint信息
func (etcd *etcdClient) connectEtcd(endPoints []string) error {
	var err error
	etcd.client, err = clientv3.New(clientv3.Config{
		Endpoints:   endPoints,
		DialTimeout: 2 * time.Second,
	})
	return err
}

//断开和etcd 服务的连接
func (etcd *etcdClient) Close() {
	etcd.client.Close()
}

//通过key值去etcd中查找对应的value
//key 为需要查找的key
//返回了类型为string
func (etcd *etcdClient) GetStringValue(key string, defaultValue string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	resp, err := etcd.client.Get(ctx, key)
	cancel()
	if err != nil {
		return ""
	}

	var value string = defaultValue
	for _, ev := range resp.Kvs {
		value = fmt.Sprintf("%s", ev.Value)
	}

	return value
}

//通过key值去etcd中查找对应的value
//key为需要查找的key
//返回类型为int
func (etcd *etcdClient) GetIntVale(key string, defaultValue int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	resp, err := etcd.client.Get(ctx, key)
	cancel()
	if err != nil {
		return 0
	}
	var value int = defaultValue
	for _, ev := range resp.Kvs {
		tmpValue := fmt.Sprintf("%d", ev.Value)
		value, _ = strconv.Atoi(tmpValue)
	}

	return value
}

//设置值64位值
func (etcd *etcdClient) SetInt64Value(key string, value int64) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	valueStr := strconv.FormatInt(value, 10)
	_, err := etcd.client.Put(ctx, key, valueStr)
	cancel()
	if err != nil {
		return R_ERR
    }
    return R_OK
}

//读取64位值
func (etcd *etcdClient) GetInt64Value(key string, defaultValue int64) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	resp, err := etcd.client.Get(ctx, key)
	cancel()
	if err != nil {
		return 0
	}
	var value int64 = defaultValue
	for _, ev := range resp.Kvs {
		tmpValue := fmt.Sprintf("%d", ev.Value)
		value, _ = strconv.ParseInt(tmpValue, 10, 64)
	}
	return value
}

//获得锁
func (etcd *etcdClient) Lock(prefixLock string) int {
	var err error
	etcd.mutex = concurrency.NewMutex(etcd.session, prefixLock)
	if etcd.mutex.Lock(context.TODO()); err != nil {
		return R_ERR
	}
	return R_OK
}

//释放锁
func (etcd *etcdClient) UnLock() int {
	if err := etcd.mutex.Unlock(context.TODO()); err != nil {
		return R_ERR
	}
	return R_OK
}