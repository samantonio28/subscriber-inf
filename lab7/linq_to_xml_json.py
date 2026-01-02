import json
import xml.etree.ElementTree as ET
from models import User
import uuid
from datetime import datetime

class LinqToXmlJson:
    def __init__(self, users):
        self.users = users
    
    def create_xml_document(self):
        """Создание XML документа"""
        root = ET.Element("Users")
        
        for user in self.users:
            user_elem = ET.SubElement(root, "User")
            ET.SubElement(user_elem, "UserId").text = str(user.user_id)
            ET.SubElement(user_elem, "Email").text = user.email
            ET.SubElement(user_elem, "UserName").text = user.user_name
            ET.SubElement(user_elem, "Age").text = str(user.age)
            ET.SubElement(user_elem, "Balance").text = str(user.balance)
        
        tree = ET.ElementTree(root)
        tree.write("users.xml", encoding="utf-8", xml_declaration=True)
        print("✓ XML документ создан: users.xml")
    
    def read_xml_document(self):
        """Чтение из XML документа"""
        tree = ET.parse("users.xml")
        root = tree.getroot()
        
        users_from_xml = []
        for user_elem in root.findall("User"):
            user_data = {
                "user_name": user_elem.find("UserName").text,
                "email": user_elem.find("Email").text,
                "balance": int(user_elem.find("Balance").text)
            }
            users_from_xml.append(user_data)
        
        print("✓ Чтение из XML:")
        for user in users_from_xml:
            print(f"  {user['user_name']} ({user['email']}): {user['balance']} руб")
        
        return users_from_xml
    
    def update_xml_document(self):
        """Обновление XML документа"""
        tree = ET.parse("users.xml")
        root = tree.getroot()
        
        for user_elem in root.findall("User"):
            if user_elem.find("UserName").text == "Иван":
                user_elem.find("Balance").text = "1200"
                print("✓ Обновлен баланс пользователя Иван")
        
        tree.write("users_updated.xml", encoding="utf-8", xml_declaration=True)
        print("✓ Обновленный XML сохранен: users_updated.xml")
    
    def add_to_xml_document(self):
        """Добавление в XML документ"""
        tree = ET.parse("users.xml")
        root = tree.getroot()
        
        new_user = ET.SubElement(root, "User")
        ET.SubElement(new_user, "UserId").text = str(uuid.uuid4())
        ET.SubElement(new_user, "Email").text = "newuser@mail.com"
        ET.SubElement(new_user, "UserName").text = "NewUser"
        ET.SubElement(new_user, "Age").text = "28"
        ET.SubElement(new_user, "Balance").text = "900"
        
        tree.write("users_extended.xml", encoding="utf-8", xml_declaration=True)
        print("✓ Новый пользователь добавлен в XML")
    
    def work_with_json(self):
        """Работа с JSON"""
        users_data = [
            {
                "user_id": str(user.user_id),
                "email": user.email,
                "user_name": user.user_name,
                "age": user.age,
                "balance": user.balance
            }
            for user in self.users
        ]
        
        # Запись в JSON
        with open("users.json", "w", encoding="utf-8") as f:
            json.dump(users_data, f, ensure_ascii=False, indent=2)
        print("✓ JSON документ создан: users.json")
        
        # Чтение из JSON
        with open("users.json", "r", encoding="utf-8") as f:
            users_from_json = json.load(f)
        
        print("✓ Чтение из JSON:")
        for user in users_from_json[:2]:  # Покажем первых двух
            print(f"  {user['user_name']}: {user['balance']} руб")
    
    def execute_all_operations(self):
        print("\n=== ЧАСТЬ 2: LINQ to XML/JSON ===")
        
        self.create_xml_document()
        self.read_xml_document()
        self.update_xml_document()
        self.add_to_xml_document()
        self.work_with_json()
