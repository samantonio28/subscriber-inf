create database rk2;
-- drop database rk2;    --     :-)

-- задание 1
create table Drivers (
	DriverID serial primary key not null,
    DriverLicense varchar(30) not null,
    FIO varchar(100) not null,
    Phone varchar(10) not null
);

create table Fines (
    FineID serial primary key not null,
    FineType int not null check(FineType > 0),
    Amount int not null check(Amount >= 0),
    FineDate date not null,
    DriverID serial not null,
    foreign key (DriverID) references Drivers(DriverID) on update cascade on delete cascade
);

create table Cars (
    CarID serial primary key not null,
    Model varchar(10) not null,
    Color varchar(10) not null,
    Year int not null check(year >= 1900),
    RegistrationDate date not null
);

create table DriversCars (
    d_id serial primary key not null,
    c_id serial not null,
    foreign key (d_id) references Drivers(DriverID) on update cascade on delete cascade,
    foreign key (c_id) references Cars(CarID) on update cascade on delete cascade
);

insert into Drivers (DriverLicense, FIO, Phone)
values 
('12340-ppp', 'abcc', '8909120'),
('12341-ppp', 'abcd', '8909121'),
('12342-ppp', 'abcf', '8909122'),
('12343-ppp', 'abcg', '8909123'),
('12344-ppp', 'abch', '8909124'),
('12345-ppp', 'abcj', '8909125'),
('12346-ppp', 'abck', '8909126'),
('12347-ppp', 'abcl', '8909127'),
('12348-ppp', 'abkm', '8909128'),
('12349-ppc', 'abcn', '8909129');

insert into Fines (FineType, Amount, FineDate, DriverID)
values
(1, 300, '2000-01-01', 1),
(1, 200, '2000-01-02', 1),
(1, 100, '2000-01-03', 2),
(1, 200, '2000-01-04', 3),
(3, 300, '2000-02-01', 4),
(4, 200, '2000-03-01', 5),
(3, 500, '2000-04-01', 6),
(3, 400, '2000-05-01', 7),
(4, 300, '2000-06-01', 8),
(4, 200, '2000-07-01', 9),
(2, 100, '2000-01-05', 1),
(2, 100, '2000-01-04', 2),
(1, 3600, '2000-01-10', 3),
(1, 400, '2000-01-11', 4),
(1, 700, '2000-01-11', 5);

insert into Cars (Model, Color, Year, RegistrationDate) 
values
('mersedes', 'r', 2000, '2000-04-03'),
('mazda', 'r', 2000, '2000-04-03'),
('mersedes', 'r', 2000, '2000-04-03'),
('mazda', 'r', 2000, '2000-04-03'),
('opel', 'r', 2000, '2000-04-03'),
('giguli', 'r', 2000, '2000-04-03'),
('giguli', 'r', 2000, '2000-04-03'),
('mersedes', 'r', 2000, '2000-04-03'),
('opel', 'r', 2000, '2000-04-03'),
('mersedes', 'r', 2000, '2000-04-03'),
('mersedes', 'r', 2000, '2000-04-03');


-- задание 2

-- находит водителя с его штрафами, указанием предыдущего и следующего штрафа для каждой строки
-- с группировкой по айди и дате штрафа (где-то разумеется будет NULL)
select 
    d.DriverID,
    FIO,
    FineDate,
    Amount,
    LAG(Amount) over (partition by d.DriverID order by FineDate) as PrevFineAmount,
    LEAD(Amount) over (partition by d.DriverID order by FineDate) as NextFineAmount
from Drivers d
join Fines f on d.DriverID = f.DriverID
order by DriverID, FineDate;

-- находит водителей у которых фио начинается на abc ИЛИ права оканчиваются на ppp (здесь напечатаются все, потому что ИЛИ)
select
    DriverID,
    FIO,
    DriverLicense,
    Phone
from Drivers
where FIO like 'abc%' 
   or DriverLicense like '%ppp';

-- находит водителей, их сумму штрафов, количество штрафов, далее пишет одну из строк
-- на основе данных о штрафах и информацию по штрафам, например что означает тип штрафа 1 (нарушение пдд) 
select 
    d.DriverID,
    d.FIO,
    count(f.FineID) as TotalFines,
    sum(f.Amount) as TotalAmount,
    case 
        when sum(f.Amount) > 1000 then 'Высокие штрафы'
        when sum(f.Amount) between 500 and 1000 then 'Средние штрафы'
        when sum(f.Amount) between 100 and 500 then 'Низкие штрафы'
        else 'Минимальные штрафы'
    end as FineCategory,
    case f.FineType
        when 1 then 'Нарушение ПДД'
        when 2 then 'Парковка'
        when 3 then 'Скорость'
        when 4 then 'Документы'
        else 'Прочее'
    end as FineTypeDescription
from Drivers d
left join Fines f on d.DriverID = f.DriverID
group by d.DriverID, d.FIO, f.FineType
order by TotalAmount desc;

-- задание 3
create or replace function find_table_with_most_constraints()
returns table (table_name text, constraint_count bigint) 
as $$
begin
    return QUERY
    select 
        tc.table_name::text,
        count(constraint_name)::bigint as constraint_count
    from information_schema.table_constraints tc
    where tc.table_schema = 'public'
    group by tc.table_name
    order by constraint_count desc
    limit 1;
end;
$$ language plpgsql;

select * from find_table_with_most_constraints();
