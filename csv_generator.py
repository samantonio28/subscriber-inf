import csv
from datetime import datetime, timedelta
import time
import os
from sqlalchemy import create_engine, text
from dataclasses import dataclass
from typing import List
import uuid
import random

@dataclass
class SubscriptionData:
    sub_id: int
    user_id: str
    service_id: int
    price: int
    sub_type: str
    start_date: str
    end_date: str
    created_at: str

class CSVDataGenerator:
    def __init__(self, connection_string: str, output_dir: str = "lab8_data/input"):
        self.engine = create_engine(connection_string)
        self.output_dir = output_dir
        os.makedirs(output_dir, exist_ok=True)
    
    def generate_file_mask(self, table_name: str) -> str:
        """Генерация имени файла по маске"""
        file_id = str(uuid.uuid4())[:8]  # Уникальный идентификатор
        current_time = datetime.now().strftime("%Y%m%d_%H%M%S")
        
        mask = f"{file_id}_{table_name}_{current_time}.csv"
        return mask
    
    def get_subscriptions_data(self, limit: int = 50) -> List[SubscriptionData]:
        """Получение данных о подписках из БД"""
        try:
            with self.engine.connect() as connection:
                query = text("""
                    SELECT 
                        sub_id, 
                        user_id, 
                        service_id, 
                        price, 
                        sub_type, 
                        start_date, 
                        end_date,
                        CURRENT_TIMESTAMP as created_at
                    FROM subscriptions 
                    ORDER BY sub_id DESC 
                    LIMIT :limit
                """)
                
                result = connection.execute(query, {"limit": limit})
                subscriptions = []
                
                for row in result:
                    subscription = SubscriptionData(
                        sub_id=row.sub_id,
                        user_id=str(row.user_id),
                        service_id=row.service_id,
                        price=row.price,
                        sub_type=row.sub_type,
                        start_date=str(row.start_date) if row.start_date else "",
                        end_date=str(row.end_date) if row.end_date else "",
                        created_at=str(row.created_at)
                    )
                    subscriptions.append(subscription)
                
                return subscriptions
                
        except Exception as e:
            print(f"Ошибка при получении данных из БД: {e}")
            return self.generate_sample_data(limit)
    
    def generate_sample_data(self, count: int = 20) -> List[SubscriptionData]:
        """Генерация тестовых данных (если в БД мало данных)"""
        subscriptions = []
        
        service_names = ["Netflix", "YouTube Premium", "Spotify", "Apple Music"]
        sub_types = ["usual", "family", "promocode"]
        
        for i in range(count):
            start_date = datetime.now() - timedelta(days=random.randint(1, 365))
            end_date = start_date + timedelta(days=random.randint(30, 365))
            
            subscription = SubscriptionData(
                sub_id=1000 + i,
                user_id=str(uuid.uuid4()),
                service_id=random.randint(1, 4),
                price=random.randint(100, 500),
                sub_type=random.choice(sub_types),
                start_date=start_date.strftime("%Y-%m-%d"),
                end_date=end_date.strftime("%Y-%m-%d"),
                created_at=datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            )
            subscriptions.append(subscription)
        
        return subscriptions
    
    def save_as_csv(self, data: List[SubscriptionData], filename: str):
        """Сохранение данных в CSV формате"""
        filepath = os.path.join(self.output_dir, filename)
        
        with open(filepath, 'w', newline='', encoding='utf-8') as f:
            if data:
                fieldnames = [field.name for field in SubscriptionData.__dataclass_fields__.values()]
                writer = csv.DictWriter(f, fieldnames=fieldnames)
                
                writer.writeheader()
                for item in data:
                    writer.writerow(item.__dict__)
        
        print(f"CSV файл создан: {filename}")
        return filepath
    
    def generate_single_file(self):
        """Генерация одного CSV файла"""
        print("=== ГЕНЕРАЦИЯ CSV ФАЙЛА ===")
        
        # Получаем данные
        data = self.get_subscriptions_data(30)
        
        if not data:
            print("Не удалось получить данные")
            return
        
        # Генерируем имя файла
        filename = self.generate_file_mask("subscriptions")
        
        # Сохраняем в CSV
        filepath = self.save_as_csv(data, filename)
        
        # Выводим статистику
        print(f"Создан файл: {filename}")
        print(f"Записей в файле: {len(data)}")
        print(f"Размер файла: {os.path.getsize(filepath)} байт")
        
        return filepath
    
    def run_continuous_generation(self, interval_minutes: int = 5):
        """Непрерывная генерация файлов с заданным интервалом"""
        print(f"=== ЗАПУСК НЕПРЕРЫВНОЙ ГЕНЕРАЦИИ CSV ФАЙЛОВ ===")
        print(f"Интервал: каждые {interval_minutes} минут")
        print(f"Папка для файлов: {self.output_dir}")
        print("Для остановки нажмите Ctrl+C\n")
        
        interval_seconds = interval_minutes * 60
        file_count = 0
        
        try:
            while True:
                file_count += 1
                print(f"\n--- Генерация файла #{file_count} ---")
                
                filepath = self.generate_single_file()
                
                if filepath:
                    print(f"Следующий файл через {interval_minutes} минут...")
                    time.sleep(interval_seconds)
                else:
                    print("Ошибка при создании файла, повтор через 1 минуту...")
                    time.sleep(60)
                    
        except KeyboardInterrupt:
            print(f"\nОстановлено. Всего создано файлов: {file_count}")
    
    def list_generated_files(self):
        """Показать список созданных файлов"""
        if not os.path.exists(self.output_dir):
            print("Папка с файлами не существует")
            return
        
        files = os.listdir(self.output_dir)
        if not files:
            print("Файлы не найдены")
            return
        
        print("=== СОЗДАННЫЕ CSV ФАЙЛЫ ===")
        for i, filename in enumerate(sorted(files), 1):
            filepath = os.path.join(self.output_dir, filename)
            file_size = os.path.getsize(filepath)
            print(f"{i}. {filename} ({file_size} байт)")

def main():
    # Настройки подключения к вашей БД
    connection_string = "postgresql://postgres:secret@localhost:8000/dev"
    
    # Создаем генератор
    generator = CSVDataGenerator(connection_string)
    
    print("CSV ГЕНЕРАТОР ДАННЫХ")
    print("1 - Создать один CSV файл")
    print("2 - Запустить непрерывную генерацию (каждые 5 минут)")
    print("3 - Показать созданные файлы")
    print("4 - Выход")
    
    while True:
        choice = input("\nВыберите действие (1-4): ").strip()
        
        if choice == "1":
            generator.generate_single_file()
        elif choice == "2":
            generator.run_continuous_generation(interval_minutes=5)
        elif choice == "3":
            generator.list_generated_files()
        elif choice == "4":
            print("Выход...")
            break
        else:
            print("Неверный выбор")

if __name__ == "__main__":
    main()
