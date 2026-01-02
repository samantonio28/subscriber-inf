from linq_promocodes import LinqPromocodeService

def main():
    connection_string = "postgresql://postgres:secret@localhost:8000/dev"
    service = LinqPromocodeService(connection_string)
    service.execute_and_save()
    
    print("--- netflix_promocodes_linq.json")
    print("--- netflix_promocodes_linq.xml")

if __name__ == "__main__":
    main()