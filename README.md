# net-multiplier
a simple tcp/udp multiplier written by golang

It duplicates a single tcp/udp input into multi outputs at the same time with independent listener and senders per task

you can dynamically add or remove task via http interface


### Theory
![Theory](./assets/net-multiplier.png)

### Usage
#### boot the application
```shell script
  -default_mode string
    	tcp or udp (default "udp")
  -local.client.host string
    	the host to which the sender is bind to (default "192.168.99.60")
  -local.http.svr.addr string
    	the http server address where the requests are handled (default "192.168.99.60:10060")
  -local.port.ceil int
    	 (default 62000)
  -local.port.floor int
    	 (default 60000)
  -local.svr.host string
    	the host to which the listener is bind to (default "192.168.99.60")
  -log.level string
    	 (default "info")
  -tempByteSliceLen int
    	the temp byte slice size for tcp/udp read (default 2048)
```
#### add task
you should send a http request to the application to add a task
```shell script
curl http://192.168.99.60/addtask?destAddrStrs=192.168.0.1:6666,192.168.0.2:8888&mode=udp
```
when the task is successfuly added,the response is like below
```json
{
    "success": true,
    "task": {
        "id": "124e4567e89b12d3a456426655440000",
        "destAddr": "192.168.99.60:62001",
        "mode": "udp"
    }
}
```

#### remove task
when you need remove the task you added before,task id is necessary 
```shell script
curl http://192.168.99.60/delTask?taskId=124e4567e89b12d3a456426655440000
```