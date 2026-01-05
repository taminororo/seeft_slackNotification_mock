import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/shift_notification.dart';

class ApiService {
  final String baseUrl;

  // 環境変数から読み込むか、デフォルト値を使用
  // Docker環境: http://go:8080
  // ローカル環境: http://localhost:8080
  ApiService({String? baseUrl})
      : baseUrl = baseUrl ??
            _getBaseUrlFromEnvironment();

  static String _getBaseUrlFromEnvironment() {
    // 環境変数から取得を試みる（Flutter Webでは実行時に取得できないため、デフォルト値を使用）
    // Docker環境では、ブラウザからアクセスするためlocalhostを使用
    // コンテナ内からアクセスする場合は、環境変数で設定
    const envBaseUrl = String.fromEnvironment('API_BASE_URL');
    if (envBaseUrl.isNotEmpty) {
      return envBaseUrl;
    }
    
    // デフォルト: ローカル開発環境
    return 'http://localhost:8080';
  }

  // 未読通知一覧を取得
 Future<List<ShiftNotification>> getNotifications(int userId) async {
  try {
    final response = await http.get(
      Uri.parse('$baseUrl/api/notifications?user_id=$userId'),
    ).timeout(const Duration(seconds: 10)); // タイムアウト追加

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      final notificationsJson = data['notifications'] as List?; // null安全
      
      if (notificationsJson == null) return [];
      
      return notificationsJson
          .map((json) => ShiftNotification.fromJson(json))
          .toList();
    } else {
      // サーバーからのエラーメッセージを含めるとデバッグしやすい
      throw Exception('ステータスコード: ${response.statusCode}');
    }
  } catch (e) {
    throw Exception('通信エラー: $e');
  }
}

  // 通知を既読にする
  Future<void> markAsRead(int notificationId, int userId) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/notifications/$notificationId/read?user_id=$userId'),
      );

      if (response.statusCode != 200) {
        throw Exception('Failed to mark as read');
      }
    } catch (e) {
      throw Exception('Error: $e');
    }
  }
}
