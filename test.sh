#!/bin/bash

# Test script for PriceMap Go project

echo "Running unit tests..."
go test ./utils -v
echo ""

echo "Running service tests..."
go test ./services -v
echo ""

echo "Running parser tests..."
go test ./parsers -v
echo ""

echo "Running API tests (short mode)..."
go test ./api -v -short
echo ""

echo "Running all tests with coverage..."
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report generated: coverage.html"

