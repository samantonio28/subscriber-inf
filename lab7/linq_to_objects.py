from models import User, Subscription
from datetime import datetime
import uuid

class LinqToObjects:
    def __init__(self):
        self.users = self._create_test_users()
        self.subscriptions = self._create_test_subscriptions()
    
    def _create_test_users(self):
        return [
            User(uuid.uuid4(), "ivanov@mail.com", "pass123", "Иван", 25, 1000),
            User(uuid.uuid4(), "petrov@mail.com", "pass456", "Петр", 30, 1500),
            User(uuid.uuid4(), "sidorov@mail.com", "pass789", "Сидор", 22, 800),
            User(uuid.uuid4(), "smirnov@mail.com", "pass000", "Смирнов", 35, 2000)
        ]
    
    def _create_test_subscriptions(self):
        return [
            Subscription(1, self.users[0].user_id, 1, 100, "usual", 
                        datetime(2024, 1, 1), datetime(2024, 2, 1)),
            Subscription(2, self.users[1].user_id, 2, 150, "family", 
                        datetime(2024, 1, 1), datetime(2024, 3, 1)),
            Subscription(3, self.users[2].user_id, 1, 120, "promocode", None, None),
            Subscription(4, self.users[3].user_id, 3, 200, "usual", 
                        datetime(2024, 2, 1), datetime(2024, 4, 1))
        ]
    
    def query_1_expensive_subscriptions(self):
        """from, where, orderby, select эквивалент"""
        result = [
            {"sub_id": sub.sub_id, "price": sub.price, "sub_type": sub.sub_type}
            for sub in self.subscriptions
            if sub.price > 100
        ]
        result.sort(key=lambda x: x["price"], reverse=True)
        return result
    
    def query_2_user_subscription_count(self):
        """join, group, into эквивалент"""
        from collections import defaultdict
        user_subs = defaultdict(list)
        
        for sub in self.subscriptions:
            user_subs[sub.user_id].append(sub)
        
        result = []
        for user in self.users:
            subscription_count = len(user_subs.get(user.user_id, []))
            result.append({"user_name": user.user_name, "subscription_count": subscription_count})
        
        return result
    
    def query_3_users_above_avg_balance(self):
        """let, where, orderby, select эквивалент"""
        avg_balance = sum(user.balance for user in self.users) / len(self.users)
        
        result = [
            {"user_name": user.user_name, "balance": user.balance, "status": "Above Average"}
            for user in self.users
            if user.balance > avg_balance
        ]
        result.sort(key=lambda x: x["balance"], reverse=True)
        return result
    
    def query_4_active_subscriptions(self):
        """from, where, orderby, select с условиями"""
        now = datetime.now()
        result = [
            {"sub_id": sub.sub_id, "start_date": sub.start_date, "end_date": sub.end_date}
            for sub in self.subscriptions
            if sub.start_date and sub.end_date and sub.end_date > now
        ]
        result.sort(key=lambda x: x["start_date"])
        return result
    
    def query_5_user_subscription_details(self):
        """join, where, orderby, select (многотабличный)"""
        user_dict = {user.user_id: user for user in self.users}
        
        result = []
        for sub in self.subscriptions:
            if sub.price > 100 and sub.user_id in user_dict:
                user = user_dict[sub.user_id]
                result.append({
                    "user_name": user.user_name,
                    "sub_type": sub.sub_type,
                    "price": sub.price,
                    "start_date": sub.start_date
                })
        
        result.sort(key=lambda x: (x["user_name"], -x["price"]))
        return result
    
    def execute_all_queries(self):
        print("=== ЧАСТЬ 1: LINQ to Objects ===")
        
        print("\n1. Дорогие подписки (>100):")
        for item in self.query_1_expensive_subscriptions():
            print(f"  Подписка {item['sub_id']}: {item['price']} руб, тип: {item['sub_type']}")
        
        print("\n2. Количество подписок по пользователям:")
        for item in self.query_2_user_subscription_count():
            print(f"  {item['user_name']}: {item['subscription_count']} подписок")
        
        print("\n3. Пользователи с балансом выше среднего:")
        for item in self.query_3_users_above_avg_balance():
            print(f"  {item['user_name']}: {item['balance']} руб - {item['status']}")
        
        print("\n4. Активные подписки:")
        for item in self.query_4_active_subscriptions():
            print(f"  Подписка {item['sub_id']}: с {item['start_date'].strftime('%d.%m.%Y')} по {item['end_date'].strftime('%d.%m.%Y')}")
        
        print("\n5. Детали подписок пользователей:")
        for item in self.query_5_user_subscription_details():
            start_date = item['start_date'].strftime('%d.%m.%Y') if item['start_date'] else 'N/A'
            print(f"  {item['user_name']}: {item['sub_type']} подписка за {item['price']} руб (с {start_date})")
