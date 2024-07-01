# Mi-Yeelight Smart Scenarios App

## A little bit smarter than just a smart 

I was faced with the problem of not being able to set the scenarios I needed due to the fact that it would not be possible to change the state of the light bulb using a timer if the light bulb was offline at the time the timer was triggered. I have implemented my scripts to fire when receiving the UDP datagram that the light bulb sends when it appears on the network. Also, every 10 minutes, scripts are checked for inclusion in the time interval and the data specified in the script is set. The project does not look ideal, but it is quite workable, at least it seems so to me.
## How to set bulb developer mode:
When the light bulb is in developer mode, it can receive commands via LAN on port 55443, this is necessary for the application to work

I've used `https://pypi.org/project/python-miio/`

Get tokens
```sh
miiocli cloud
``` 

Set "Developer mode"

```sh
miiocli yeelight --ip 192.168.1.51 --token e587ee623b383cf7f9515308305839d2 set_developer_mode 1
```

## How to configure

in the `config` directory there is a yaml configuration file `scenarios.yaml`, this is an example of a script, I won’t describe it, it’s easy to understand

## How to run
`main.go` is located in the `cmd` directory

Something like this will work:
```sh
go run cmd/main.go

```
