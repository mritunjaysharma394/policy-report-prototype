# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM alpine:latest
COPY kube-bench kube-bench
COPY policyreport policyreport

# Command to run the executable
CMD ["/policyreport"]