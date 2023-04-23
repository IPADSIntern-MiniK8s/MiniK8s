package storage

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
)

type MyStruct struct {
	Field1 string
	Field2 int
}

func TestStorage(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2380"},
	})
	if err != nil {
		return
	}

	// test create
	myStruct := &MyStruct{Field1: "Hello", Field2: 42}
	etcdStorage := NewEtcdStorage(client)
	err = etcdStorage.Create(context.Background(), "myStruct", &myStruct)
	var expectedErr error = nil

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	myStruct5 := &MyStruct{Field1: "Hello2", Field2: 43}
	err = etcdStorage.Create(context.Background(), "myStruct2", &myStruct5)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// test get
	var myStruct2 MyStruct
	err = etcdStorage.Get(context.Background(), "myStruct", &myStruct2)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if myStruct2.Field1 != myStruct.Field1 {
		t.Errorf("Expected %v, got %v", myStruct.Field1, myStruct2.Field1)
	}
	if myStruct2.Field2 != myStruct.Field2 {
		t.Errorf("Expected %v, got %v", myStruct.Field2, myStruct2.Field2)
	}

	// test getList
	var myStructList []MyStruct
	err = etcdStorage.GetList(context.Background(), "myStruct", &myStructList)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// test update
	myStruct.Field1 = "World"
	err = etcdStorage.GuaranteedUpdate(context.Background(), "myStruct", &myStruct)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// test get
	var myStruct3 MyStruct
	err = etcdStorage.Get(context.Background(), "myStruct", &myStruct3)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if myStruct3.Field1 != myStruct.Field1 {
		t.Errorf("Expected %v, got %v", myStruct.Field1, myStruct3.Field1)
	}

	// test delete
	err = etcdStorage.Delete(context.Background(), "myStruct")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// test get
	var myStruct4 MyStruct
	err = etcdStorage.Get(context.Background(), "myStruct", &myStruct4)
	if err == expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	var ErrKeyNotFound = fmt.Errorf("key not found: myStruct")
	if err != ErrKeyNotFound {
		fmt.Print(ErrKeyNotFound)
	}
}
