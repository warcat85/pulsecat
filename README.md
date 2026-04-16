# PulseCat - Системный мониторинг

## Общее описание
Демон - программа, собирающая информацию о системе, на которой запущена,
и отправляющая её своим клиентам по GRPC.

## Архитектура
- GRPC сервер;
- допускается использование временных (`/tmp`) файлов;
- статистика хранится в памяти, долговременное хранение не предусмотрено.

## Требования
Необходимо каждые **N** секунд выдавать информацию, усредненную за последние **M** секунд.

Например, N = 5с, а M = 15с, тогда демон "молчит" первые 15 секунд,
затем выдает снапшот за 0-15с; через 5с (в 20с) выдает снапшот за 5-20с;
через 5с (в 25с) выдает снапшот за 10-25с и т.д.

**N** и **M** указывает клиент в запросе на получение статистики.

Что необходимо собирать:
- Средняя загрузка системы (load average).

- Средняя загрузка CPU (%user_mode, %system_mode, %idle).

- Загрузка дисков:
    - tps (transfers per second);
    - KB/s (kilobytes (read+write) per second);

- Информация о дисках по каждой файловой системе:
    - использовано мегабайт, % от доступного количества;
    - использовано inode, % от доступного количества.

- Top talkers по сети:
    - по протоколам: protocol (TCP, UDP, ICMP, etc), bytes, % от sum(bytes) за последние **M**), сортируем по убыванию процента;
    - по трафику: source ip:port, destination ip:port, protocol, bytes per second (bps), сортируем по убыванию bps.

- Статистика по сетевым соединениям:
    - слушающие TCP & UDP сокеты: command, pid, user, protocol, port;
    - количество TCP соединений, находящихся в разных состояниях (ESTAB, FIN_WAIT, SYN_RCV и пр.).

#### Разрешено использовать только стандартную библиотеку языка Go!

Команды, которые могут пригодиться:
```
$ top -b -n1
$ df -k
$ df -i
$ iostat -d -k
$ cat /proc/net/dev
$ sudo netstat -lntup
$ ss -ta
$ tcpdump -ntq -i any -P inout -l
$ tcpdump -nt -i any -P inout -ttt -l
```

Статистика представляет собой объекты, описанные в формате Protobuf.

Информацию необходимо выдавать всем подключенным по GRPC клиентам
с использованием [однонаправленного потока](https://grpc.io/docs/tutorials/basic/go/#server-side-streaming-rpc).

Выдавать "снапшот" системы можно как отдельными сообщениями, так и одним жирным объектом.

Сбор информации, её парсинг и пр. должен осуществляться как можно более конкурентно.

## Поддерживаемая ОС
Минимум - Linux (Ubuntu 18.04).

Максимум - несколько сборок под набор из популярных ОС/процессоров:
- darwin, linux, windows
- 386, amd64

[Список возможных вариантов](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63).

Но тогда придется постараться с реализацией использования различных команд для сбора данных.

Пригодятся [build тэги](https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures).

## Конфигурация
- Через аргументы командной строки можно указать, на каком порту стартует сервер.
- Через файл можно указать, какие из подсистем сбора включены/выключены.

## Тестирование
#### Юнит-тесты
- по возможности мок интерфейсов и проверка вызовов конкретных методов;
- тесты вспомогательных функций и пр.

#### Интеграционные тесты
- потестировать факт потока статистики, можно без конкретных цифр;
- можно посоздавать файлы, пооткрывать сокеты и посмотреть на изменение снапшота.

#### Клиент
Необходимо реализовать простой клиент, который в реальном времени получает
и выводит в STDOUT статистику по одному из пунктов (например, сетевую информацию)
в читаемом формате (например, в виде таблицы).

Проект включает клиент **PulseKitten**, который подключается к серверу PulseCat и отображает статистику в реальном времени.

##### Использование PulseKitten

Соберите клиент:
```bash
make build
```
или
```bash
go build -o pulsekitten ./cmd/pulsekitten
```

Запустите клиент с указанием адреса сервера:
```bash
./pulsekitten -server localhost:25225 -frequency 5 -stats load,cpu,disk
```

Доступные параметры:
- `-server`: адрес сервера PulseCat (по умолчанию localhost:25225)
- `-delay`: начальная задержка в секундах (M параметр)
- `-frequency`: частота снимков в секундах (N параметр)
- `-stats`: фильтр типов статистики (load,cpu,disk,network,talkers,sockets,tcp)
- `-meow`: отправить meow-запрос вместо подписки на статистику (нельзя использовать вместе с -stats)
- `-duration`: продолжительность работы в секундах (0 = бесконечно)
- `-verbose`: подробный вывод
- `-version`: показать версию

Пример вывода:
```
=== System Stats at 2026-04-16T15:33:43+03:00 ===
Load Average: 1m=0.23, 5m=0.21, 15m=0.18
CPU Usage: User=13.5%, System=8.2%, Idle=71.3%, Nice=0.0%, IOWait=0.0%
Disk Usage (1 filesystems):
  /: 50.3% used (51243 MB used / 102400 MB total)
Network: RX=1043000 bytes, TX=543000 bytes
```

Пример использования meow-запроса:
```bash
./pulsekitten -server localhost:25225 --meow -delay 2 -frequency 3
```

Вывод meow-запроса:
```
Meow #1 at 2026-04-16T15:35:12+03:00: Meow!
Meow #2 at 2026-04-16T15:35:15+03:00: Meow!
Meow #3 at 2026-04-16T15:35:18+03:00: Meow!
```

## Разбалловка
Максимум - **20 баллов**
(при условии выполнения [обязательных требований](./README.md)):

* Реализован сбор:
    - load average - 1 балл;
    - загрузка CPU - 1 балл;
    - загрузка дисков - 1 балл;
    - top talkers по сети - 1 балла;
    - статистика по сети - 1 балл.
* Через конфигурацию можно отключать отдельную статистику - 2 балла.
* Написаны юнит-тесты - 1 балл.
* Написаны интеграционные тесты - 2 балла.
* Реализован простой клиент к демону - 2 балла.
* Сбор хотя бы одного типа статистики работает на разных ОС - 5 баллов.
* Понятность и чистота кода - до 3 баллов.

#### Зачёт от 10 баллов

## Docker Deployment

The system monitor daemon can be run in a Docker container for easy deployment.

### Building the Docker Image

```bash
docker build -t pulsecat:latest .
```

### Running with Docker

For basic operation (limited monitoring capabilities):
```bash
docker run -d \
  --name pulsecat \
  -p 25225:25225 \
  pulsecat:latest
```

For full system monitoring capabilities (recommended):
```bash
docker run -d \
  --name pulsecat \
  --privileged \
  --network host \
  -v /proc:/host/proc:ro \
  -v /sys:/host/sys:ro \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v $(pwd)/configs/config.yaml:/configs/config.yaml:ro \
  pulsecat:latest \
  --config configs/config.yaml
```

### Using Docker Compose

A `compose.yaml` file is provided for easier deployment:

```bash
# Start the service
docker compose up -d

# View logs
docker compose logs -f

# Stop the service
docker compose down
```

### Configuration

The Docker image does not include a configuration file by default. The `compose.yaml` mounts `./configs/config.yaml` from the host into the container. If you don't have a configuration file, copy the example:

```bash
cp configs/config.yaml.example configs/config.yaml
```

You can customize the configuration by editing `configs/config.yaml`. To override with your own configuration file, modify the volume mount in `compose.yaml` or use a custom Docker run command:

```bash
docker run -d \
  --name pulsecat \
  -v /path/to/your/config.yaml:/configs/config.yaml:ro \
  pulsecat:latest \
  --config configs/config.yaml
```

### Notes

1. **Privileged Mode**: System monitoring requires access to host system information. Running with `--privileged` and mounting `/proc` and `/sys` is necessary for full functionality.

2. **Network Mode**: Using `--network host` provides accurate network statistics. If port mapping is preferred, use `-p 25225:25225` instead.

3. **Health Checks**: The container includes a health check that verifies the gRPC server is listening on port 25225.

4. **Security**: For production deployments, consider using more restrictive capabilities instead of full privileged mode, and run as a non-root user when possible.
