from sqlalchemy import create_engine, Column, Integer, String, DateTime, ForeignKey, func, and_
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime, timedelta
import json
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from typing import List, Dict, Any

Base = declarative_base()

class ServiceDB(Base):
    __tablename__ = 'services'
    
    service_id = Column(Integer, primary_key=True)
    service_name = Column(String)
    sub_duration_id_default = Column(Integer)
    users_count = Column(Integer)
    has_promocodes = Column(Integer)

class UserDB(Base):
    __tablename__ = 'users'
    
    user_id = Column(String, primary_key=True)
    email = Column(String)
    password = Column(String(30))
    user_name = Column(String(20))
    age = Column(Integer)
    balance = Column(Integer)

class SubscriptionDB(Base):
    __tablename__ = 'subscriptions'
    
    sub_id = Column(Integer, primary_key=True)
    user_id = Column(String, ForeignKey('users.user_id'))
    service_id = Column(Integer, ForeignKey('services.service_id'))
    price = Column(Integer)
    sub_type = Column(String)
    start_date = Column(DateTime)
    end_date = Column(DateTime)
    
    user = relationship("UserDB", backref="subscriptions")
    service = relationship("ServiceDB", backref="subscriptions")

class PromocodeDB(Base):
    __tablename__ = 'promocodes'
    
    promocode_id = Column(Integer, primary_key=True)
    service_id = Column(Integer, ForeignKey('services.service_id'))
    promocode = Column(String(10))
    sub_duration_days = Column(Integer)
    sub_id = Column(Integer, ForeignKey('subscriptions.sub_id'))
    expires_at = Column(DateTime)
    
    service = relationship("ServiceDB", backref="promocodes")
    subscription = relationship("SubscriptionDB", backref="promocodes")

@dataclass
class PromocodeInfo:
    service_name: str
    user_name: str
    user_email: str
    promocode: str
    expires_at: datetime
    subscription_end_date: datetime
    days_until_expiry: int

class LinqPromocodeService:
    def __init__(self, connection_string: str):
        self.engine = create_engine(connection_string)
        self.Session = sessionmaker(bind=self.engine)
    
    def get_urgency_level(self, days: int) -> str:
        if days <= 1:
            return "CRITICAL"
        elif days <= 3:
            return "HIGH"
        elif days <= 5:
            return "MEDIUM"
        else:
            return "LOW"
    
    def generate_message(self, record: PromocodeInfo) -> str:
        return (f"У {record.user_name} подписка истекает через {record.days_until_expiry} дней. "
                f"Промокод {record.promocode} действителен до {record.expires_at.strftime('%Y-%m-%d')}")
    
    def execute_netflix_promocodes_linq(self) -> List[PromocodeInfo]:        
        session = self.Session()
        try:
            current_date = datetime.now().date()
            end_date_threshold = current_date + timedelta(days=70)
            
            # linq запрос
            query = (
                session.query(
                    ServiceDB.service_name,
                    UserDB.user_name,
                    UserDB.email,
                    PromocodeDB.promocode,
                    PromocodeDB.expires_at,
                    SubscriptionDB.end_date,
                    (func.date(SubscriptionDB.end_date) - func.current_date()).label('days_until_expiry')
                )
                .join(ServiceDB, PromocodeDB.service_id == ServiceDB.service_id)
                .join(SubscriptionDB, PromocodeDB.sub_id == SubscriptionDB.sub_id)
                .join(UserDB, SubscriptionDB.user_id == UserDB.user_id)
                .filter(ServiceDB.service_name == 'Netflix')
                .filter(PromocodeDB.expires_at >= current_date)
                .filter(SubscriptionDB.end_date.between(current_date, end_date_threshold))
                .filter(SubscriptionDB.end_date.isnot(None))
                .order_by('days_until_expiry')
                .order_by(PromocodeDB.expires_at)
            )
            
            results = []
            for row in query:
                promocode_info = PromocodeInfo(
                    service_name=row.service_name,
                    user_name=row.user_name,
                    user_email=row.email,
                    promocode=row.promocode,
                    expires_at=row.expires_at,
                    subscription_end_date=row.end_date,
                    days_until_expiry=row.days_until_expiry
                )
                results.append(promocode_info)
            
            return results
            
        except Exception as e:
            print(f"Ошибка при выполнении LINQ запроса: {e}")
            return []
        finally:
            session.close()
    
    def save_to_json(self, records: List[PromocodeInfo], filename: str = "netflix_promocodes_linq.json"):
        formatted_results = []
        for record in records:
            formatted_results.append({
                "service_name": record.service_name,
                "user_name": record.user_name,
                "user_email": record.user_email,
                "promocode": record.promocode,
                "promocode_expires": record.expires_at.strftime("%Y-%m-%d"),
                "subscription_ends": record.subscription_end_date.strftime("%Y-%m-%d"),
                "days_until_expiry": record.days_until_expiry,
                "urgency_level": self.get_urgency_level(record.days_until_expiry),
                "message": self.generate_message(record),
            })
        
        output = {
            "query_type": "netflix_promocodes_linq",
            "description": "Действительные промокоды Netflix для пользователей с истекающими подписками (LINQ версия)",
            "results": formatted_results,
            "summary": {
                "total_promocodes": len(records),
            }
        }
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(output, f, ensure_ascii=False, indent=2)
        
        print(f"JSON документ создан: {filename}")
        return output
    
    def save_to_xml(self, records: List[PromocodeInfo], filename: str = "netflix_promocodes_linq.xml"):        
        root = ET.Element("NetflixPromocodes")
        
        summary = ET.SubElement(root, "Summary")
        ET.SubElement(summary, "TotalPromocodes").text = str(len(records))
        ET.SubElement(summary, "QueryMethod").text = "LINQ"
        
        promocodes_elem = ET.SubElement(root, "Promocodes")
        for record in records:
            promocode_elem = ET.SubElement(promocodes_elem, "Promocode")
            
            ET.SubElement(promocode_elem, "ServiceName").text = record.service_name
            ET.SubElement(promocode_elem, "UserName").text = record.user_name
            ET.SubElement(promocode_elem, "UserEmail").text = record.user_email
            ET.SubElement(promocode_elem, "PromocodeValue").text = record.promocode
            ET.SubElement(promocode_elem, "PromocodeExpires").text = record.expires_at.strftime("%Y-%m-%d")
            ET.SubElement(promocode_elem, "SubscriptionEnds").text = record.subscription_end_date.strftime("%Y-%m-%d")
            ET.SubElement(promocode_elem, "DaysUntilExpiry").text = str(record.days_until_expiry)
            ET.SubElement(promocode_elem, "UrgencyLevel").text = self.get_urgency_level(record.days_until_expiry)
        
        tree = ET.ElementTree(root)
        tree.write(filename, encoding='utf-8', xml_declaration=True)
        print(f"XML документ создан: {filename}")
    
    def execute_and_save(self):
        print("linq")
        records = self.execute_netflix_promocodes_linq()
        
        if not records:
            print("Нет данных для сохранения")
            return
        
        print(f"Найдено {len(records)} записей")
        print("2. в json")
        self.save_to_json(records)
        print("3. в xml")
        self.save_to_xml(records)
        
        return records
