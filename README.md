# livekit-record-minio

This repo contains code for: 

1. Configuring a `livekit` egress server so that it saves recordings to a `minio` instance.
2. Adding a simple inotify hook that will upload recordings to a nextcloud instance using `webdav` and share it with an arbitrary entity in Nextcloud (A `talk` conversation is used by default).

## How to use

First, compile any of the two binaries that you think you're going to need:

```
go build ./cmd/livekit-minio
go build ./cmd/minio-hook
```

## configuration

### minio upload

The first binary you need to place in the box that you will use to control
recordings. You might want to dockerize it and make it available from wherever 
your authentication app lives.

Edit the provided sample config:

```
cp livekit-minio-example.env .env
```

### minio webdav hook

Here you might want to implement a different uploader, but one is provided for convenience.

On the folder/container where you want to run the inotify watcher, do:

```
cp mino-hook-example.env .env
```

The minio instance is local in this case; you need to provide also nextcloud credentials
to do the webdav upload and the sharing with the destination group. by default,
we're using a Talk ID, but this can be customized to your needs.

## running

On the API server:

```
❯ ./livekit-minio

   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v3.3.10-dev
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:3000
```

On the minio vm:

```
./minio-hook
```

## Control start & stop 

The control utility offers two endpoints. You need, obviously, to pass credentials to connect to your egress server. It expects two parameters:

* the room name
* the sharewith entity. this will be encoded in the filename, and later used to known with which nextcloud entity the audiobot needs to share the file.

An example will probably make it clearer. To start recording:

```
curl "localhost:3000/start?room=testroom&shareWith=a92tv9pf"
```

And to stop it:

```
curl "localhost:3000/stop?room=testroom
```

Please note that stopping the service will make it lose state (i.e., will lose any ongoing recording), since the live recordings are only persisted in memory so far.

Beware also that there's no authentication implemented in the control plane; it's up to you to check that the calls to make recordings are properly checked for ACLs and sanitized.

If you're using django, you might want to use [django-livekit-api](https://github.com/kalikaneko/django-livekit-api) to schedule livekit rooms and to add permissions for joining and recording.

## Minio installation & config 

Follow instructions to install [minio](https://min.io/docs/minio/linux/operations/install-deploy-manage/deploy-minio-single-node-single-drive.html#minio-snsd)

```
wget https://dl.min.io/server/minio/release/linux-amd64/archive/minio_20231223071911.0.0_amd64.deb -O minio.deb
sudo dpkg -i minio.deb
```

1. Make sure that the service is up and running, and create a bucket (`livekit-uploads`).
2. Create access key (take note of the access key and secret).
3. Create a policy for the bucket

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "arn:aws:s3:::livekit-recordings/*"
            ]
        }
    ]
}
```
4. Create user `livekit` and assign the livekit policy.
