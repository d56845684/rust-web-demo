docker build -t rust-demo .
docker image prune -f
docker run --rm -p 8080:8080 -v /Users/dennis/rust/rust-web-demo/static:/app/static --name rust-demo rust-demo