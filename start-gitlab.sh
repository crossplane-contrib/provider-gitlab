docker run \
  --publish 8443:443 --publish 8088:80 --publish 8022:22 \
  --name gitlab \
  --restart always \
  --shm-size 256m \
  gitlab/gitlab-ce:latest