FROM ubuntu:22.04

# 環境変数の設定
ENV DEBIAN_FRONTEND=noninteractive
ENV FLUTTER_VERSION=3.16.0
ENV PATH="/opt/flutter/bin:${PATH}"

# 必要なパッケージのインストール
RUN apt-get update && apt-get install -y \
    curl \
    git \
    unzip \
    xz-utils \
    zip \
    libglu1-mesa \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Flutter SDKのインストール
RUN git clone --branch ${FLUTTER_VERSION} https://github.com/flutter/flutter.git /opt/flutter

# Flutter doctorの実行（ライセンス承認とセットアップ）
RUN flutter doctor || true
RUN flutter config --enable-web

# 作業ディレクトリの設定
WORKDIR /app

# pubspec.yamlをコピーして依存関係をインストール
COPY pubspec.yaml ./
RUN flutter pub get

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 3000

# Flutter Webサーバーを起動（ホットリロード有効）
CMD ["flutter", "run", "-d", "web-server", "--web-port=3000", "--web-hostname=0.0.0.0", "--web-renderer=html"]

