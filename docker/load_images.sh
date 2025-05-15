#!/bin/sh
echo "Loading Docker images from /tar_files..."

for tarfile in /tar_files/*.tar; do
  echo "Loading: $tarfile"
  docker load -i "$tarfile"
done

echo "All images are loaded."