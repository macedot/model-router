#!/bin/bash
set -e
for script in test/test-*.sh; do
	echo "Running $script..."
	bash "$script"
done
