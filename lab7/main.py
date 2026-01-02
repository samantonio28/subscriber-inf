from linq_to_objects import LinqToObjects
from linq_to_xml_json import LinqToXmlJson
from linq_to_sql import LinqToSql

def main():
    print("=== ЛАБОРАТОРНАЯ РАБОТА №7 - Python реализация ===")
    
    # Часть 1: LINQ to Objects
    linq_objects = LinqToObjects()
    linq_objects.execute_all_queries()
    
    # Часть 2: LINQ to XML/JSON
    linq_xml_json = LinqToXmlJson(linq_objects.users)
    linq_xml_json.execute_all_operations()
    
    # Часть 3: LINQ to SQL
    connection_string = "postgresql://postgres:secret@localhost:8000/dev"
    try:
        linq_sql = LinqToSql(connection_string)
        linq_sql.execute_all_operations()
    except Exception as e:
        print(f"\nЧасть 3 (LINQ to SQL) пропущена: {e}")
        print("Убедитесь, что база данных запущена и строка подключения верна")
    
    print("\nЛабораторная работа завершена!")

if __name__ == "__main__":
    main()
