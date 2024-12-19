#!/bin/bash
set -e

# Compilar el binario
go build -o build/usr/local/bin/webalert-service main.go

# Crear el paquete Debian
dpkg-deb --build build webalert-service-latest.deb