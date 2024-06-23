package main

/*
   GKSE - Go Kea Stats Exporter
   Copyright (C) 2023 Tobias Klausmann

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; version 2 of the License ONLY.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License along
   with this program; if not, write to the Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "0.1.0"

var (
	listen   = flag.String("l", ":9988", "IP:port to listen on")
	timeout  = flag.Duration("timeout", time.Second*3, "Timeout for webserver reading client request")
	logColor = flag.Bool("cl", false, "Enable color in logs")

	logger *slog.Logger
)

func main() {
	flag.Parse()
	logger = logSetup(os.Stderr, slog.LevelInfo, "20060102-15:04:05.000", *logColor)

	logger.Info("Kea DHCP v4 stats exporter starting", "version", version)
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:              *listen,
		ReadHeaderTimeout: *timeout,
	}

	reg := prometheus.NewPedanticRegistry()
	jt := newKeaCollector(*namespace)
	prometheus.MustRegister(jt, reg)
	logger.Info("Starting webserver", "listenAddress", *listen)
	logger.Error("Exiting", "reason", srv.ListenAndServe())
}
