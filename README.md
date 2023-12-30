# requirements

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
