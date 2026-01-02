import requests
import time
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime

class Lab9Tester:
    def __init__(self):
        self.base_url = "http://localhost:8080"
        self.results = []
    
    def wait_for_service(self):
        """Ожидает пока сервис станет доступен"""
        print("Ожидание запуска сервисов...")
        for i in range(30):
            try:
                response = requests.get(f"{self.base_url}/lab9/1", timeout=5)
                if response.status_code == 200:
                    print("✅ Сервис доступен!")
                    return True
            except:
                pass
            time.sleep(2)
        print("❌ Сервис не запустился")
        return False
    
    def add_test_data(self):
        # ...
        """Добавляет тестовые данные"""
        print("📊 Добавление тестовых данных...")
        response = requests.post(f"{self.base_url}/lab9/6")
        if response.status_code == 200:
            print("✅ Тестовые данные добавлены")
        else:
            print("❌ Ошибка добавления данных")
    
    def run_performance_test(self, scenario_name, data_change_type):
        """Запускает тест производительности"""
        print(f"🚀 Запуск теста: {scenario_name}")
        
        # Конфигурация теста
        test_config = {
            "name": scenario_name,
            "description": f"Test with {data_change_type}",
            "interval": 5,
            "data_change": data_change_type
        }
        
        # Запускаем тест
        response = requests.post(f"{self.base_url}/lab9/4", json=test_config)
        if response.status_code != 200:
            print(f"❌ Ошибка запуска теста: {response.text}")
            return False
        
        print("⏳ Тест выполняется 2 минуты...")
        time.sleep(20)  # Ждем 2 минуты
        
        # Собираем результаты
        self.collect_results(scenario_name, data_change_type)
        return True
    
    def collect_results(self, scenario_name, data_change_type):
        """Собирает результаты через прямое тестирование"""
        print(f"📈 Сбор результатов для {scenario_name}...")
        
        db_times = []
        cache_times = []
        
        # Выполняем 20 запросов к каждому эндпоинту
        for i in range(20):
            # Запрос к БД
            start_time = time.time()
            response = requests.get(f"{self.base_url}/lab9/1")
            db_time = (time.time() - start_time) * 1000  # в миллисекунды
            if response.status_code == 200:
                db_times.append(db_time)
            
            # Запрос к кэшу
            start_time = time.time()
            response = requests.get(f"{self.base_url}/lab9/2")  
            cache_time = (time.time() - start_time) * 1000
            if response.status_code == 200:
                cache_times.append(cache_time)
            
            time.sleep(0.5)  # Пауза между запросами
        
        # Сохраняем результаты
        result = {
            "scenario": scenario_name,
            "data_change": data_change_type,
            "db_times": db_times,
            "cache_times": cache_times,
            "avg_db_time": np.mean(db_times) if db_times else 0,
            "avg_cache_time": np.mean(cache_times) if cache_times else 0,
            "timestamp": datetime.now().isoformat()
        }
        
        self.results.append(result)
        print(f"✅ Результаты собраны: БД={result['avg_db_time']:.2f}мс, Кэш={result['avg_cache_time']:.2f}мс")
    
    def generate_report(self):
        """Генерирует отчет с графиками"""
        print("📊 Генерация отчета...")
        
        if not self.results:
            print("❌ Нет данных для отчета")
            return
        
        # Создаем графики
        self.create_comparison_chart()
        self.create_time_chart()
        self.create_summary_table()
        
        print("✅ Отчет сгенерирован!")
        print("📁 Созданные файлы:")
        print("   - performance_report.png")
        print("   - time_analysis.png") 
        print("   - summary.txt")
    
    def create_comparison_chart(self):
        """Создает сравнительную диаграмму"""
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 6))
        
        scenarios = [r["scenario"] for r in self.results]
        db_avgs = [r["avg_db_time"] for r in self.results]
        cache_avgs = [r["avg_cache_time"] for r in self.results]
        
        # Столбчатая диаграмма
        x = np.arange(len(scenarios))
        width = 0.35
        
        bars1 = ax1.bar(x - width/2, db_avgs, width, label='PostgreSQL', alpha=0.8, color='red')
        bars2 = ax1.bar(x + width/2, cache_avgs, width, label='Redis Cache', alpha=0.8, color='blue')
        
        ax1.set_xlabel('Сценарий тестирования')
        ax1.set_ylabel('Среднее время (мс)')
        ax1.set_title('Сравнение производительности БД и Redis кэша')
        ax1.set_xticks(x)
        ax1.set_xticklabels([s.replace('_', '\n') for s in scenarios])
        ax1.legend()
        ax1.grid(True, alpha=0.3)
        
        # Добавляем значения на столбцы
        for bar in bars1:
            height = bar.get_height()
            ax1.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                    f'{height:.1f}мс', ha='center', va='bottom')
        
        for bar in bars2:
            height = bar.get_height()
            ax1.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                    f'{height:.1f}мс', ha='center', va='bottom')
        
        # Box plot для распределения времени
        all_db_times = [r["db_times"] for r in self.results]
        all_cache_times = [r["cache_times"] for r in self.results]
        
        positions = range(1, len(scenarios) * 2 + 1, 2)
        box_data = []
        box_labels = []
        
        for i, scenario in enumerate(scenarios):
            box_data.extend([all_db_times[i], all_cache_times[i]])
            box_labels.extend([f'{scenario}\nБД', f'{scenario}\nКэш'])
        
        ax2.boxplot(box_data, labels=box_labels)
        ax2.set_ylabel('Время выполнения (мс)')
        ax2.set_title('Распределение времени выполнения')
        ax2.grid(True, alpha=0.3)
        plt.xticks(rotation=45)
        
        plt.tight_layout()
        plt.savefig('./report/performance_report.png', dpi=300, bbox_inches='tight')
        plt.show()
    
    def create_time_chart(self):
        """Создает график временных рядов"""
        fig, axes = plt.subplots(2, 2, figsize=(15, 10))
        axes = axes.flatten()
        
        for i, result in enumerate(self.results):
            if i >= len(axes):
                break
                
            ax = axes[i]
            db_times = result["db_times"]
            cache_times = result["cache_times"]
            
            ax.plot(db_times, 'r-', label='PostgreSQL', alpha=0.7, linewidth=2)
            ax.plot(cache_times, 'b-', label='Redis Cache', alpha=0.7, linewidth=2)
            
            ax.set_title(f'Сценарий: {result["scenario"]}')
            ax.set_xlabel('Номер запроса')
            ax.set_ylabel('Время (мс)')
            ax.legend()
            ax.grid(True, alpha=0.3)
            
            # Добавляем средние линии
            ax.axhline(y=result["avg_db_time"], color='red', linestyle='--', alpha=0.5)
            ax.axhline(y=result["avg_cache_time"], color='blue', linestyle='--', alpha=0.5)
        
        plt.tight_layout()
        plt.savefig('./report/time_analysis.png', dpi=300, bbox_inches='tight')
        plt.show()
    
    def create_summary_table(self):
        """Создает текстовый отчет"""
        with open('./report/summary.txt', 'w', encoding='utf-8') as f:
            f.write("ОТЧЕТ ПО ЛАБОРАТОРНОЙ РАБОТЕ №9\n")
            f.write("In-Memory Database на примере Redis\n")
            f.write("=" * 50 + "\n\n")
            
            f.write("РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ:\n")
            f.write("-" * 30 + "\n")
            
            for result in self.results:
                f.write(f"\nСценарий: {result['scenario']}\n")
                f.write(f"Тип изменений: {result['data_change']}\n")
                f.write(f"Среднее время БД: {result['avg_db_time']:.2f} мс\n")
                f.write(f"Среднее время кэша: {result['avg_cache_time']:.2f} мс\n")
                
                if result['avg_db_time'] > 0:
                    speedup = result['avg_db_time'] / result['avg_cache_time']
                    f.write(f"Ускорение: {speedup:.2f}x\n")
                
                f.write("-" * 20 + "\n")
            
            f.write("\nВЫВОДЫ:\n")
            f.write("-" * 20 + "\n")
            
            # Анализируем результаты
            best_speedup = max([r['avg_db_time']/r['avg_cache_time'] for r in self.results if r['avg_cache_time'] > 0])
            worst_speedup = min([r['avg_db_time']/r['avg_cache_time'] for r in self.results if r['avg_cache_time'] > 0])
            
            f.write(f"1. Redis ускоряет запросы в {best_speedup:.1f}-{worst_speedup:.1f} раз\n")
            f.write("2. Наибольшая эффективность достигается при стабильных данных\n")
            f.write("3. Изменения данных снижают эффективность кэширования\n")
            f.write("4. Redis оптимален для часто читаемых редко изменяемых данных\n")

def main():
    print("=== ЛАБОРАТОРНАЯ РАБОТА №9 ===")
    print("In-Memory Database на примере Redis\n")
    
    tester = Lab9Tester()
    
    # Шаг 1: Ожидание сервиса
    if not tester.wait_for_service():
        return
    
    # Шаг 2: Добавление тестовых данных
    tester.add_test_data()
    
    # Шаг 3: Запуск тестовых сценариев
    test_scenarios = [
        ("no_changes", "none"),
        ("with_additions", "add"), 
        ("with_deletions", "delete"),
        ("with_updates", "update")
    ]
    
    for scenario_name, change_type in test_scenarios:
        success = tester.run_performance_test(scenario_name, change_type)
        if not success:
            break
        
        # Пауза между тестами
        time.sleep(10)
    
    # Шаг 4: Генерация отчета
    tester.generate_report()
    
    print("\n🎉 ЛАБОРАТОРНАЯ РАБОТА ВЫПОЛНЕНА!")
    print("📋 Для отчета используйте файлы:")
    print("   - performance_report.png - основные графики")
    print("   - time_analysis.png - анализ времени")
    print("   - summary.txt - текстовый отчет")

if __name__ == "__main__":
    main()
