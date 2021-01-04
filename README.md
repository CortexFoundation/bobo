# Bobo

## Simplest web server (less than ```70``` lines coding of Golang) but available

Implement the basic features of a web backend with the simplest ways. The orignal only has ```bobo.go``` for simple, it is ok to change or split it as needed

Feel free to fork it and add advanced logics and features

## How to run it ?
```
go run bobo.go
```
or 
```
go build bobo.go
./bobo
```
## How to use it ?
```
time ./bench.sh

... ...

real	0m29.865s
user	0m13.490s
sys	0m3.074s
```

#### Storage
https://github.com/dgraph-io/badger

## Customized
## Users
### 1. Register
To register user information in json format
##### Method
```
POST
```
##### URL
```
/user/0x970e8128ab834e8eac17ab8e3812f010678cf791?sig=0x15ce17f60e6825a4d5556867c30d3bc823f9f2dd0d55aa845a816f4518a081ca5e2c9fea9ec552e861d015306c6c7c4132135e97b0e695e01c751c51e5e7075d01
```
##### Params
```
sig: the signature of user info in DATA below {...}
```

##### DATA

ts : current timestamp used for signature checking
```
"{\"ts\":1609310997, \"name\":\"jo\"}"
```

### 2. Details
To find user information by address (```0x970e8128ab834e8eac17ab8e3812f010678cf791```)
##### Method
```
GET
```
##### URL
```
/user/0x970e8128ab834e8eac17ab8e3812f010678cf791
```
##### Params
```
NULL
```

##### DATA
```
NULL
```
## Works
### 1. Favor
To Favor the work
##### Method
```
POST
```
##### URL
```
/favor/0x970e8128ab834e8eac17ab8e3812f010678cf791?sig=0xab133a9294a829e9023264cd146a8ab56d9ff8ead591cc7d218d5cbd1513f8a73fe7c6666b86f8aaffef8ba6f3cef4bfaebe4a7502df052803965440da7baa7300
```
##### Params
```
sig: the signature of user info in DATA below {...}
```

##### DATA

ts : current timestamp used for signature checking
```
"{\"ts\":1609310997, \"addr\":\"0x2a2a0667f9cbf4055e48eaf0d5b40304b8822184\"}"
```

### 2. Favored
To find user information by address (```0x970e8128ab834e8eac17ab8e3812f010678cf791```)
##### Method
```
GET
```
##### URL
```
/favored/0x970e8128ab834e8eac17ab8e3812f010678cf791
```
##### Params
```
NULL
```

##### DATA
```
NULL
```

test domain http://share.cortexlabs.ai:8081
