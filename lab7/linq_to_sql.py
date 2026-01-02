from sqlalchemy import create_engine, Column, Integer, String, DateTime, ForeignKey, Text, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
import uuid
from datetime import datetime

Base = declarative_base()

class UserDB(Base):
    __tablename__ = 'users'
    
    user_id = Column(Text, primary_key=True)
    email = Column(String)
    password = Column(String(30))
    user_name = Column(String(20))
    age = Column(Integer)
    balance = Column(Integer)

class SubscriptionDB(Base):
    __tablename__ = 'subscriptions'
    
    sub_id = Column(Integer, primary_key=True)
    user_id = Column(Text, ForeignKey('users.user_id'))
    service_id = Column(Integer)
    price = Column(Integer)
    sub_type = Column(String)
    start_date = Column(DateTime)
    end_date = Column(DateTime)

class LinqToSql:
    def __init__(self, connection_string):
        self.engine = create_engine(connection_string)
        self.Session = sessionmaker(bind=self.engine)
    
    def single_table_select(self):
        """Однотабличный запрос на выборку"""
        session = self.Session()
        try:
            # Эквивалент LINQ: from u in users where u.balance > 10500 select u
            users = session.query(UserDB).filter(UserDB.balance > 10500).order_by(UserDB.balance.desc()).all()
            
            print("✓ Однотабличный запрос - пользователи с балансом > 10500:")
            for user in users:
                print(f"  {user.user_name}: {user.balance} руб")
            
            return users
        finally:
            session.close()
    
    def multi_table_select(self):
        """Многотабличный запрос на выборку"""
        session = self.Session()
        try:
            # Эквивалент LINQ join
            results = session.query(UserDB.user_name, SubscriptionDB.sub_type, SubscriptionDB.price)\
                .join(SubscriptionDB, UserDB.user_id == SubscriptionDB.user_id)\
                .filter(SubscriptionDB.price < 500)\
                .order_by(UserDB.user_name, SubscriptionDB.price.desc())\
                .all()
            
            print("✓ Многотабличный запрос - подписки дешевле 500 руб:")
            for user_name, sub_type, price in results:
                print(f"  {user_name}: {sub_type} за {price} руб")
            
            return results
        finally:
            session.close()
    
    def crud_operations(self):
        """Операции добавления, изменения, удаления"""
        session = self.Session()
        try:
            # Добавление
            new_user = UserDB(
                user_id=str(uuid.uuid4()),
                email="test@mail.com",
                password="testpass",
                user_name="TestUser",
                age=26,
                balance=750
            )
            session.add(new_user)
            print("✓ Пользователь добавлен")
            
            # Изменение
            user_to_update = session.query(UserDB).filter_by(user_name="TestUser").first()
            if user_to_update:
                user_to_update.balance = 850
                print("✓ Баланс пользователя обновлен")
            
            # Удаление
            user_to_delete = session.query(UserDB).filter_by(user_name="TestUser").first()
            if user_to_delete:
                session.delete(user_to_delete)
                print("✓ Пользователь удален")
            
            session.commit()
        except Exception as e:
            session.rollback()
            print(f"✗ Ошибка: {e}")
        finally:
            session.close()
    
    def execute_stored_procedure(self):
        """Выполнение хранимой процедуры"""
        session = self.Session()
        try:
            # Создадим тестовую хранимую процедуру если её нет
            create_proc_sql = """
            CREATE OR REPLACE FUNCTION get_high_balance_users(min_balance INTEGER)
            RETURNS TABLE(user_name VARCHAR, balance INTEGER) AS $$
            BEGIN
                RETURN QUERY SELECT u.user_name, u.balance 
                            FROM users u 
                            WHERE u.balance >= min_balance
                            ORDER BY u.balance DESC;
            END; $$
            LANGUAGE plpgsql;
            """
            session.execute(text(create_proc_sql))
            session.commit()
            
            # Вызов хранимой процедуры
            result = session.execute(text("SELECT * FROM get_high_balance_users(:min_balance)"), 
                                   {"min_balance": 10000})
            
            print("✓ Хранимая процедура - пользователи с высоким балансом:")
            for row in result:
                print(f"  {row.user_name}: {row.balance} руб")
                
        except Exception as e:
            print(f"✗ Ошибка хранимой процедуры: {e}")
        finally:
            session.close()
    
    def execute_all_operations(self):
        print("\n=== ЧАСТЬ 3: LINQ to SQL (SQLAlchemy) ===")
        
        self.single_table_select()
        self.multi_table_select()
        self.crud_operations()
        self.execute_stored_procedure()
