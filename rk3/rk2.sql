-- create database rk3;

create table driver (
    DriverID serial primary key not null,
    FIO varchar(100) not null,
    BirthDate date not null,
    OnboardedAt date not null,
    Region varchar(10) not null 
);

create table route (
    DriverID serial not null,
    RaceDate date not null,
    RaceTime timestamp not null, 
    WeekDay varchar(13) not null,
    route_type boolean not null
);

insert into driver (FIO, BirthDate, OnboardedAt, Region)
values 
('ИИИ', '1990-10-20', '2024-02-02', 'Москва'),
('ППП', '2000-05-15', '2024-02-10', 'Москва'),
('CCC', '2005-03-03', '2024-03-15', 'К обл');

insert into route (DriverID, RaceDate, RaceTime, WeekDay, route_type)
values
(1, '2025-10-20', '2001-09-28 14:00'::timestamp, 'Понедельник', false),
(1, '2025-10-21', '2001-09-28 10:00'::timestamp, 'Вторник', true),
(1, '2025-10-21', '2001-09-28 18:15'::timestamp, 'Вторник', false),
(2, '2025-10-20', '2001-09-28 14:00'::timestamp, 'Понедельник', false),
(1, '2025-10-25', '2001-09-28 12:04'::timestamp, 'Четверг', true);

-- Задание 1
-- 1 маршруты, когда водитель приехал, которые встречаются меньше двух раз

-- 2 второй запрос - запрос на поиск водителей с минимальным стажем

