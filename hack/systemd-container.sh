docker run -it --rm \
  --mount type=bind,source=/sys/fs/cgroup,target=/sys/fs/cgroup \
  --mount type=tmpfs,destination=/tmp \
  --mount type=tmpfs,destination=/run \
  --mount type=tmpfs,destination=/run/lock \
  -e USER=admin -e ENABLE_ROOT=true --cap-add CAP_SYS_ADMIN \
    quay.io/tinyzimmer/kvdi:arch-kde-0.0.1
