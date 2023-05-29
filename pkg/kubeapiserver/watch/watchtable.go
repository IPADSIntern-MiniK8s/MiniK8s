package watch

import "sync"

// WatchTable map the attribute name to the watch server
var WatchTable = make(map[string]*WatchServer)

var WatchStorage sync.Map
