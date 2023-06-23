package limiter

import "sync"

// GlobalAgentLock is a lock that exists to prevent multiple requests performing multiple Docker calls at the same time,
// the reasoning for this is that each Docker call (create image, create volume, create container, remove container, etc)
// have a wide-range of effects to different services.
//
// As such, we want to avoid causing random side effects, to do that, we limit our agent to handle one request
// at a time and since we are using server-side events, the clients should be able to wait until all the requests
// are done.
var GlobalAgentLock = sync.Mutex{}
