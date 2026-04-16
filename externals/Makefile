.PHONY: build_external_local

ROOT_DIR := $(abspath $(CURDIR)/..)
GO_MODULE_MODE ?= vendor
EXTERNAL_NAME ?=
EXTERNAL_GOOS ?= linux
EXTERNAL_ARCH ?= $(shell go env GOARCH)
EXTERNAL_OUTDIR ?=
EXTERNAL_OUTPUT ?=
EXTERNAL_LDFLAGS ?= -w -s
EXTERNAL_EBPF_ARGS ?=
DK_BPF_KERNEL_SRC_PATH ?=
EXTERNAL_SYSROOT_BASE ?= /opt/sysroots/debian10

build_external_local:
	@test -n "$(EXTERNAL_NAME)" || { echo "EXTERNAL_NAME is required"; exit 1; }
	@output_name="$(EXTERNAL_OUTPUT)"; \
	if [ -z "$$output_name" ]; then \
		output_name="$(EXTERNAL_NAME)"; \
	fi; \
	outdir="$(EXTERNAL_OUTDIR)"; \
	if [ -z "$$outdir" ]; then \
		outdir="$(ROOT_DIR)/dist/datakit-$(EXTERNAL_GOOS)-$(EXTERNAL_ARCH)/externals"; \
	fi; \
	case "$$outdir" in \
		/*) outdir_abs="$$outdir" ;; \
		*) outdir_abs="$(ROOT_DIR)/$$outdir" ;; \
	esac; \
	host_arch="$$(uname -m)"; \
	target_triplet=""; \
	cc="gcc"; \
	cxx="g++"; \
	case "$(EXTERNAL_ARCH)" in \
		amd64) \
			target_triplet="x86_64-linux-gnu"; \
			if command -v x86_64-linux-gnu-gcc >/dev/null 2>&1; then \
				cc="x86_64-linux-gnu-gcc"; \
				cxx="x86_64-linux-gnu-g++"; \
			elif [ "$$host_arch" != "x86_64" ]; then \
				echo "missing x86_64-linux-gnu cross toolchain in $$host_arch build environment"; \
				exit 1; \
			fi; \
			;; \
		arm64) \
			target_triplet="aarch64-linux-gnu"; \
			if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then \
				cc="aarch64-linux-gnu-gcc"; \
				cxx="aarch64-linux-gnu-g++"; \
			elif [ "$$host_arch" != "aarch64" ]; then \
				echo "missing aarch64-linux-gnu cross toolchain in $$host_arch build environment"; \
				exit 1; \
			fi; \
			;; \
	esac; \
	export CC="$$cc" CXX="$$cxx"; \
	sysroot="$(EXTERNAL_SYSROOT_BASE)/$(EXTERNAL_ARCH)"; \
	if [ -n "$$target_triplet" ] && [ -d "$$sysroot" ]; then \
		export SYSROOT="$$sysroot"; \
		export PKG_CONFIG_SYSROOT_DIR="$$SYSROOT"; \
		export PKG_CONFIG_LIBDIR="$$SYSROOT/usr/lib/$$target_triplet/pkgconfig:$$SYSROOT/usr/lib/pkgconfig:$$SYSROOT/usr/share/pkgconfig"; \
		export PKG_CONFIG_PATH=; \
		export PKG_CONFIG_ALLOW_CROSS=1; \
		export CGO_CPPFLAGS="--sysroot=$$SYSROOT"; \
		export CGO_CFLAGS="--sysroot=$$SYSROOT -Wno-undef-prefix"; \
		export CGO_CXXFLAGS="--sysroot=$$SYSROOT"; \
		export CGO_LDFLAGS="--sysroot=$$SYSROOT -L$$SYSROOT/lib/$$target_triplet -L$$SYSROOT/usr/lib/$$target_triplet -Wl,-rpath-link,$$SYSROOT/lib/$$target_triplet -Wl,-rpath-link,$$SYSROOT/usr/lib/$$target_triplet"; \
	else \
		export CGO_CFLAGS="-Wno-undef-prefix"; \
	fi; \
	mkdir -p "$$outdir_abs"; \
	case "$(EXTERNAL_NAME)" in \
		ebpf) \
			kernel_src="$(DK_BPF_KERNEL_SRC_PATH)"; \
			if [ -z "$$kernel_src" ]; then \
				kernel_src="/usr/src/linux-headers-$(EXTERNAL_ARCH)"; \
			fi; \
			$(MAKE) --no-print-directory -C "$(ROOT_DIR)/internal/plugins/externals/ebpf" \
				SRCPATH=. \
				OUTPATH="$$outdir_abs/$$output_name" \
				ARCH="$(EXTERNAL_ARCH)" \
				GOOS="$(EXTERNAL_GOOS)" \
				GOARCH="$(EXTERNAL_ARCH)" \
				CGO_ENABLED=1 \
				DK_BPF_KERNEL_SRC_PATH="$$kernel_src" \
				ARGS="$(EXTERNAL_EBPF_ARGS)"; \
			;; \
		db2|journald|oracle) \
			GO111MODULE=on GOFLAGS=-mod=$(GO_MODULE_MODE) CGO_ENABLED=1 \
				GOOS="$(EXTERNAL_GOOS)" GOARCH="$(EXTERNAL_ARCH)" \
				go build -tags netgo -buildvcs=false \
				-ldflags "$(EXTERNAL_LDFLAGS)" \
				-o "$$outdir_abs/$$output_name" \
				"$(ROOT_DIR)/internal/plugins/externals/$(EXTERNAL_NAME)/main.go"; \
			;; \
		logfwd) \
			GO111MODULE=on GOFLAGS=-mod=$(GO_MODULE_MODE) CGO_ENABLED=0 \
				GOOS="$(EXTERNAL_GOOS)" GOARCH="$(EXTERNAL_ARCH)" \
				go build -buildvcs=false \
				-ldflags "$(EXTERNAL_LDFLAGS)" \
				-o "$$outdir_abs/$$output_name" \
				"$(ROOT_DIR)/internal/plugins/externals/logfwd/cmd/main.go"; \
			;; \
		*) \
			GO111MODULE=on GOFLAGS=-mod=$(GO_MODULE_MODE) CGO_ENABLED=0 \
				GOOS="$(EXTERNAL_GOOS)" GOARCH="$(EXTERNAL_ARCH)" \
				go build -buildvcs=false \
				-ldflags "$(EXTERNAL_LDFLAGS)" \
				-o "$$outdir_abs/$$output_name" \
				"$(ROOT_DIR)/internal/plugins/externals/$(EXTERNAL_NAME)/main.go"; \
			;; \
	esac
