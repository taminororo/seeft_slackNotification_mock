import 'package:flutter/material.dart';
import '../models/shift_notification.dart';
import '../services/api_service.dart';

class NotificationListScreen extends StatefulWidget {
  final int userId;
  final ApiService apiService;

  const NotificationListScreen({
    Key? key,
    required this.userId,
    required this.apiService,
  }) : super(key: key);

  @override
  State<NotificationListScreen> createState() => _NotificationListScreenState();
}

class _NotificationListScreenState extends State<NotificationListScreen> {
  List<ShiftNotification> _notifications = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadNotifications();
  }

  Future<void> _loadNotifications() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final notifications = await widget.apiService.getNotifications(widget.userId);
      setState(() {
        _notifications = notifications;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('通知の取得に失敗しました: $e')),
        );
      }
    }
  }

Future<void> _openNotificationDetail(ShiftNotification notification) async {
    // 1. 【楽観的更新】APIの結果を待たずに、まずUI上の既読状態をtrueにする
    setState(() {
      final index = _notifications.indexWhere((n) => n.id == notification.id);
      if (index != -1) {
        _notifications[index] = ShiftNotification(
          id: notification.id,
          userName: notification.userName,
          yearId: notification.yearId,
          timeId: notification.timeId,
          date: notification.date,
          weather: notification.weather,
          oldTaskName: notification.oldTaskName,
          newTaskName: notification.newTaskName,
          isRead: true, // 即座に既読にする
          createdAt: notification.createdAt,
        );
      }
    });

    // 2. 詳細画面を表示（APIの完了を待たずに遷移）
    if (mounted) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => NotificationDetailScreen(notification: notification),
        ),
      );
    }

    // 3. 【バックグラウンド実行】裏側でAPIを呼び出す
    try {
      // ここで await しても、既に画面遷移とUI更新は終わっているのでユーザーを待たせない
      await widget.apiService.markAsRead(notification.id, widget.userId);
    } catch (e) {
      // 4. 【ロールバック】APIが失敗した場合は未読に戻す
      if (mounted) {
        setState(() {
          final index = _notifications.indexWhere((n) => n.id == notification.id);
          if (index != -1) {
            _notifications[index] = ShiftNotification(
              id: notification.id,
              userName: notification.userName,
              yearId: notification.yearId,
              timeId: notification.timeId,
              date: notification.date,
              weather: notification.weather,
              oldTaskName: notification.oldTaskName,
              newTaskName: notification.newTaskName,
              isRead: false, // 未読に戻す
              createdAt: notification.createdAt,
            );
          }
        });
        // ユーザーに失敗を通知
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('同期に失敗しました。未読に戻します: $e')),
        );
      }
    }
  }

  String _timeIdToTime(int timeId) {
    // timeID 25 = 6:00 を基準とする
    final hoursFromBase = (timeId - 25) ~/ 2;
    final minutesFromBase = ((timeId - 25) % 2) * 30;
    final hours = 6 + hoursFromBase;
    final minutes = minutesFromBase;
    return '${hours.toString().padLeft(2, '0')}:${minutes.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('シフト変更通知'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadNotifications,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _notifications.isEmpty
              ? const Center(child: Text('通知はありません'))
              : RefreshIndicator(
                  onRefresh: _loadNotifications,
                  child: ListView.builder(
                    itemCount: _notifications.length,
                    itemBuilder: (context, index) {
                      final notification = _notifications[index];
                      return Card(
                        margin: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 8,
                        ),
                        // 未読の場合は枠色を強調
                        color: notification.isRead
                            ? Colors.white
                            : Colors.blue.shade50,
                        elevation: notification.isRead ? 2 : 4,
                        child: ListTile(
                          title: Text(
                            '${notification.userName} - ${notification.date}',
                            style: TextStyle(
                              fontWeight: notification.isRead
                                  ? FontWeight.normal
                                  : FontWeight.bold,
                            ),
                          ),
                          subtitle: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text('時刻: ${_timeIdToTime(notification.timeId)}'),
                              if (notification.oldTaskName.isNotEmpty)
                                Text(
                                  '変更前: ${notification.oldTaskName}',
                                  style: const TextStyle(
                                    decoration: TextDecoration.lineThrough,
                                    color: Colors.grey,
                                  ),
                                ),
                              Text('変更後: ${notification.newTaskName}'),
                            ],
                          ),
                          trailing: notification.isRead
                              ? const Icon(Icons.check_circle, color: Colors.grey)
                              : const Icon(Icons.circle, color: Colors.blue),
                          onTap: () => _openNotificationDetail(notification),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}

class NotificationDetailScreen extends StatelessWidget {
  final ShiftNotification notification;

  const NotificationDetailScreen({
    Key? key,
    required this.notification,
  }) : super(key: key);

  String _timeIdToTime(int timeId) {
    final hoursFromBase = (timeId - 25) ~/ 2;
    final minutesFromBase = ((timeId - 25) % 2) * 30;
    final hours = 6 + hoursFromBase;
    final minutes = minutesFromBase;
    return '${hours.toString().padLeft(2, '0')}:${minutes.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('通知詳細'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildDetailRow('ユーザー', notification.userName),
            _buildDetailRow('日付', notification.date),
            _buildDetailRow('天気', notification.weather),
            _buildDetailRow('時刻', _timeIdToTime(notification.timeId)),
            if (notification.oldTaskName.isNotEmpty)
              _buildDetailRow(
                '変更前',
                notification.oldTaskName,
                textStyle: const TextStyle(
                  decoration: TextDecoration.lineThrough,
                  color: Colors.grey,
                ),
              ),
            _buildDetailRow('変更後', notification.newTaskName),
            _buildDetailRow('作成日時', notification.createdAt),
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value, {TextStyle? textStyle}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              label,
              style: const TextStyle(
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: textStyle,
            ),
          ),
        ],
      ),
    );
  }
}

