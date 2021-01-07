#!/bin/bash

tmux new-session 'kubectl proxy' \; \
    select-pane -t 0 \; \
    split-window -h 'kubectl port-forward --pod-running-timeout=6h svc/prometheus-service -n monitoring 8080:8080' \; \
    select-pane -t 0 \; \
    split-window -v 'kubectl port-forward pods/monitoring-grafana-0 -n monitoring 2000:3000' \; \
    select-pane -t 2 \; \
    split-window -v 'kubectl port-forward services/mysql -n monitoring 3306:3306' \; \
    attach