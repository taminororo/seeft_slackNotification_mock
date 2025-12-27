class ShiftNotification {
  final int id;
  final String userName;
  final int yearId;
  final int timeId;
  final String date;
  final String weather;
  final String oldTaskName;
  final String newTaskName;
  final bool isRead;
  final String createdAt;

  ShiftNotification({
    required this.id,
    required this.userName,
    required this.yearId,
    required this.timeId,
    required this.date,
    required this.weather,
    required this.oldTaskName,
    required this.newTaskName,
    required this.isRead,
    required this.createdAt,
  });

  factory ShiftNotification.fromJson(Map<String, dynamic> json) {
    return ShiftNotification(
      id: json['id'] as int,
      userName: json['user_name'] as String,
      yearId: json['year_id'] as int,
      timeId: json['time_id'] as int,
      date: json['date'] as String,
      weather: json['weather'] as String,
      oldTaskName: json['old_task_name'] as String? ?? '',
      newTaskName: json['new_task_name'] as String,
      isRead: json['is_read'] as bool,
      createdAt: json['created_at'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_name': userName,
      'year_id': yearId,
      'time_id': timeId,
      'date': date,
      'weather': weather,
      'old_task_name': oldTaskName,
      'new_task_name': newTaskName,
      'is_read': isRead,
      'created_at': createdAt,
    };
  }
}

