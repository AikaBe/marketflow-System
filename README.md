    hello

    Что бы все открылось сначало нужно написать вот так:

    chmod +x load_images.sh

    потом вот так:

    docker-compose up load_images

    и только потом так:

     docker-compose up

     хай, мой варик:

# 1. Перейти в корень проекта (где docker-compose.yml)
cd marketflow

# 2. Загрузить .tar-образы из ./docker/tar_files
docker-compose run --rm load_images

# 3. Запустить все сервисы, включая marketflow и exchange1/2/3
docker-compose up --build
