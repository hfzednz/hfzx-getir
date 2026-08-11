enum DutyStatus {
  offline,
  online,
  busy,
  onBreak,
  emergency;

  String get apiValue => switch (this) {
        DutyStatus.offline => 'offline',
        DutyStatus.online => 'online',
        DutyStatus.busy => 'busy',
        DutyStatus.onBreak => 'on_break',
        DutyStatus.emergency => 'emergency',
      };

  static DutyStatus fromApi(String? value) {
    switch (value) {
      case 'online':
        return DutyStatus.online;
      case 'busy':
        return DutyStatus.busy;
      case 'on_break':
      case 'break':
        return DutyStatus.onBreak;
      case 'emergency':
        return DutyStatus.emergency;
      case 'offline':
      default:
        return DutyStatus.offline;
    }
  }
}
