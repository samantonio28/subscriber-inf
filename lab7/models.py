from dataclasses import dataclass
from datetime import datetime
from typing import Optional, List
import uuid

@dataclass
class User:
    user_id: uuid.UUID
    email: str
    password: str
    user_name: str
    age: int
    balance: int

@dataclass
class Subscription:
    sub_id: int
    user_id: uuid.UUID
    service_id: int
    price: int
    sub_type: str
    start_date: Optional[datetime]
    end_date: Optional[datetime]

@dataclass
class Service:
    service_id: int
    service_name: str
    sub_duration_id_default: int
    users_count: int
    has_promocodes: bool
