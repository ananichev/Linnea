# Linnea

### What is this?

Linnea is a easy to use image uploader.

### Setup

[Here](./docs/setup.md)

### Docker

```bash
make docker
docker run -p 8080:8080 --restart=unless-stopped -v $PWD/config.json:/app/config.json joachimflottorp/linnea
```
