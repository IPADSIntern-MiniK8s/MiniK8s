package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
)

type EtcdStorage struct {
	client *clientv3.Client
}

func (e *EtcdStorage) Client() *clientv3.Client {
	return e.client
}

func NewEtcdStorage(client *clientv3.Client) *EtcdStorage {
	return &EtcdStorage{client: client}
}

// Get retrieves the value at the specified key.
// the interface description in k8s.io/apiserver/pkg/storage/interfaces.go:
// Get unmarshals json found at key into objPtr. On a not found error, will either
// return a zero apiobject of the requested type, or an error, depending on 'opts.ignoreNotFound'.
// Treats empty responses and nil response nodes exactly like a not found error.
// The returned contents may be delayed, but it is guaranteed that they will
// match 'opts.ResourceVersion' according 'opts.ResourceVersionMatch'.
func (e *EtcdStorage) Get(ctx context.Context, key string, out interface{}) error {
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return err
	}
	if resp.Kvs == nil || len(resp.Kvs) == 0 {
		return fmt.Errorf("key not found: %s", key)
	}
	if err := json.Unmarshal(resp.Kvs[0].Value, out); err != nil {
		return err
	}
	return nil
}

// Create creates a new key with the given value.
// TODO：need to consider TTL ?
// the interface description in k8s.io/apiserver/pkg/storage/interfaces.go
// Create adds a new apiobject at a key unless it already exists. 'ttl' is time-to-live
// in seconds (0 means forever). If no error is returned and out is not nil, out will be
// set to the read value from database.
func (e *EtcdStorage) Create(ctx context.Context, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = e.client.Put(ctx, key, string(jsonValue))
	return err
}

// Delete removes the specified key.
// the interface description in k8s.io/apiserver/pkg/storage/interfaces.go:
// If key didn't exist, it will return NotFound storage error.
// If 'cachedExistingObject' is non-nil, it can be used as a suggestion about the
// current version of the apiobject to avoid read operation from storage to get it.
// However, the implementations have to retry in case suggestion is stale.
func (e *EtcdStorage) Delete(ctx context.Context, key string) error {
	_, err := e.client.Delete(ctx, key)
	return err
}

// Watch begins watching the specified key. The watch interface returned sends events
// to the returned channel. The provided context controls the entire watch lifecycle.
// The channel is closed when the context is canceled or when the server returns a
// non-retryable error. The provided revision is used as a starting point for the watch.
func (e *EtcdStorage) Watch(ctx context.Context, key string, callback func(string, []byte) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := e.client.Watch(ctx, key, clientv3.WithPrefix())

	for {
		select {
		case wresp := <-ch:
			for _, ev := range wresp.Events {
				err := callback(string(ev.Kv.Key), ev.Kv.Value)
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// GuaranteedUpdate keeps trying to update key 'key' (of type 'ptrToType')
// retrying the update until success if there is index conflict.
func (e *EtcdStorage) GuaranteedUpdate(ctx context.Context, key string, newData interface{}) error {
	for {
		// Get the current version of the data to update
		var existingData interface{}
		resp, err := e.client.Get(ctx, key)
		if err != nil {
			return err
		}
		if resp.Kvs == nil || len(resp.Kvs) == 0 {
			return fmt.Errorf("key not found: %s", key)
		}
		if err := json.Unmarshal(resp.Kvs[0].Value, &existingData); err != nil {
			return err
		}

		// Compare the current data to the new data to see if an update is needed
		if existingData == newData {
			return nil // No update needed
		}

		// Create a transaction to update the data with optimistic concurrency control
		jsonValue, err := json.Marshal(newData)
		if err != nil {
			return err
		}
		txn := e.Client().Txn(ctx).
			If(clientv3.Compare(clientv3.ModifiedRevision(key), "=", resp.Kvs[0].ModRevision)).
			Then(clientv3.OpPut(key, string(jsonValue)))

		// Execute the transaction
		txnResp, err := txn.Commit()
		if err != nil {
			return err
		}

		// Check if the transaction succeeded
		if !txnResp.Succeeded {
			// Another client updated the data, so retry the update
			continue
		}

		// The update succeeded, so return
		return nil
	}
}

// Close closes the storage client.
func (e *EtcdStorage) Close() error {
	return e.client.Close()
}
