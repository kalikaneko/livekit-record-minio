# livekit-record-minio

This repo contains code to configure `livekit` egress server to save recordings to a `minio` instance, 
and a simple inotify hook that will upload recordings to a nextcloud instance using `webdav`.

## how to use

Compile the two binaries:

```
go build ./cmd/livekit-minio
go build ./cmd/minio-webdav
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

```
cp mino-webdav-example.env .env
```

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

## Minio installation & config 

Follow instructions to install [https://min.io/docs/minio/linux/operations/install-deploy-manage/deploy-minio-single-node-single-drive.html#minio-snsd](minio):

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
