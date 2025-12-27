load('ext://restart_process', 'docker_build_with_restart')
load('ext://uibutton', 'cmd_button')

# NETWORK MONITOR

local_resource(
  'network-monitor-compile',
  'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/network-monitor-dev ./cmd/network_monitor/main.go',
  labels=["Compile"],
  deps=['./cmd/network_monitor/main.go', './internal/'])

docker_build_with_restart(
  'network-monitor-image',
  '.',
  dockerfile="dev/network-monitor.Dockerfile",
  entrypoint=['/app/build/network-monitor-dev', '--ping-ips=8.8.8.8,79.79.79.79', '--log-level=DEBUG'],
  only=[ './build'],
  live_update=[
    sync('./build/network-monitor-dev', '/app/build/network-monitor-dev'),
  ])

k8s_yaml('dev/network-monitor.k8s.yaml')
k8s_resource('network-monitor', port_forwards=8080, labels=["Binaries"])

# GO TESTS RESOURCE

local_resource(
    'go-tests',
    cmd='go test ./...',
    labels=["Tests"],
    trigger_mode=TRIGGER_MODE_MANUAL,
    auto_init=False
)

cmd_button(
    name='run-go-tests',
    resource='go-tests',
    argv=['go', 'test', './...'],
    text='Run Go Tests',
    icon_name='check_circle'
)

# PROMETHEUS SERVER

CONFIG_PATH = "dev/prometheus.yml"
ABSOLUTE_DIR = os.path.dirname(__file__)
HOST_CONFIG_PATH = os.path.join(ABSOLUTE_DIR, CONFIG_PATH)

local_resource(
    "prometheus-server",
    serve_cmd="""
        docker run --rm \
            --name prometheus \
            -p 9090:9090 \
            --mount type=bind,source=%s,target=/etc/prometheus/prometheus.yml \
            prom/prometheus
    """ % (HOST_CONFIG_PATH),
    deps=[CONFIG_PATH],
    labels=["Prometheus"])
