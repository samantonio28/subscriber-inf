# ========== КОНФИГУРАЦИЯ ==========
# DB_CONFIG = {
#     'host': 'localhost',      # или ваш хост
#     'port': 8100,
#     'database': 'dev',
#     'user': 'postgres',
#     'password': 'secret'
# }

#!/usr/bin/env python3
"""
Генератор ER-диаграммы в нотации Чена для yEd Graph Editor
"""

import psycopg2
from psycopg2.extras import DictCursor
import xml.etree.ElementTree as ET
from datetime import datetime
import math
import argparse

# ========== КОНФИГУРАЦИЯ ==========
DB_CONFIG = {
    'host': 'localhost',      # или ваш хост
    'port': 8100,
    'database': 'dev',
    'user': 'postgres',
    'password': 'secret'
}

# ========== ФУНКЦИИ ДЛЯ РАБОТЫ С БД ==========

def get_db_schema():
    """Получает структуру БД"""
    conn = psycopg2.connect(**DB_CONFIG)
    cursor = conn.cursor(cursor_factory=DictCursor)
    
    # Получаем все таблицы пользовательской схемы
    cursor.execute("""
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_type = 'BASE TABLE'
        ORDER BY table_name;
    """)
    tables = [row['table_name'] for row in cursor.fetchall()]
    
    schema_info = {}
    
    for table in tables:
        # Получаем колонки
        cursor.execute("""
            SELECT 
                column_name,
                data_type,
                is_nullable
            FROM information_schema.columns 
            WHERE table_name = %s 
            AND table_schema = 'public'
            ORDER BY ordinal_position;
        """, (table,))
        columns = [dict(row) for row in cursor.fetchall()]
        
        # Получаем PRIMARY KEY
        cursor.execute("""
            SELECT kcu.column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu 
                ON tc.constraint_name = kcu.constraint_name
            WHERE tc.constraint_type = 'PRIMARY KEY'
            AND tc.table_name = %s
            AND tc.table_schema = 'public';
        """, (table,))
        pk_columns = [row['column_name'] for row in cursor.fetchall()]
        
        # Получаем FOREIGN KEY
        cursor.execute("""
            SELECT
                kcu.column_name,
                ccu.table_name AS foreign_table_name,
                ccu.column_name AS foreign_column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu
                ON tc.constraint_name = kcu.constraint_name
            JOIN information_schema.constraint_column_usage ccu
                ON ccu.constraint_name = tc.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_name = %s
            AND tc.table_schema = 'public';
        """, (table,))
        foreign_keys = [dict(row) for row in cursor.fetchall()]
        
        schema_info[table] = {
            'columns': columns,
            'primary_keys': pk_columns,
            'foreign_keys': foreign_keys
        }
    
    cursor.close()
    conn.close()
    return schema_info

def get_relationships(schema_info):
    """Определяет связи между сущностями"""
    relationships = []
    seen = set()
    
    for table, info in schema_info.items():
        for fk in info['foreign_keys']:
            rel_key = f"{table}_{fk['foreign_table_name']}"
            if rel_key not in seen:
                seen.add(rel_key)
                relationships.append({
                    'from_entity': table,
                    'from_column': fk['column_name'],
                    'to_entity': fk['foreign_table_name'],
                    'to_column': fk['foreign_column_name'],
                    'name': f"{table}_to_{fk['foreign_table_name']}"
                })
    
    return relationships

# ========== ГЕНЕРАЦИЯ VALID GRAPHML ДЛЯ yEd ==========

def generate_valid_graphml(schema_info, relationships, output_file='er_diagram.graphml'):
    """Генерирует валидный GraphML файл для yEd"""
    
    # Создаём корневой элемент с правильными пространствами имён
    graphml = ET.Element('graphml')
    graphml.set('xmlns', 'http://graphml.graphdrawing.org/xmlns')
    graphml.set('xmlns:y', 'http://www.yworks.com/xml/graphml')
    graphml.set('xmlns:xsi', 'http://www.w3.org/2001/XMLSchema-instance')
    graphml.set('xsi:schemaLocation', 
                'http://graphml.graphdrawing.org/xmlns http://www.yworks.com/xml/schema/graphml/1.1/ygraphml.xsd')
    
    # Добавляем необходимые ключи
    key_d6 = ET.SubElement(graphml, 'key', id='d6', yfiles__type='nodegraphics')
    key_d10 = ET.SubElement(graphml, 'key', id='d10', yfiles__type='edgegraphics')
    
    # Создаём граф
    graph = ET.SubElement(graphml, 'graph', id='G', edgedefault='directed')
    
    entities = list(schema_info.keys())
    node_counter = 0
    all_nodes = {}
    
    # Координатная сетка
    cols = max(3, int(math.ceil(len(entities) ** 0.5)))
    start_x, start_y = 100, 100
    step_x, step_y = 400, 300
    
    # ===== 1. Сущности (прямоугольники) =====
    for idx, entity in enumerate(entities):
        node_id = f'n{node_counter}'
        all_nodes[f"entity_{entity}"] = node_id
        node_counter += 1
        
        row = idx // cols
        col = idx % cols
        x = start_x + col * step_x
        y = start_y + row * step_y
        
        node = ET.SubElement(graph, 'node', id=node_id)
        data = ET.SubElement(node, 'data', key='d6')
        
        # GenericNode для сущности (прямоугольник)
        generic_node = ET.SubElement(data, 'y:GenericNode', 
                                     configuration='com.yworks.entityRelationship.small_entity')
        
        geometry = ET.SubElement(generic_node, 'y:Geometry', 
                                 height='50.0', width='120.0', 
                                 x=str(x), y=str(y))
        
        fill = ET.SubElement(generic_node, 'y:Fill', 
                            color='#FFFFFF', color2='#FFFFFF', transparent='false')
        
        border = ET.SubElement(generic_node, 'y:BorderStyle', 
                              color='#000000', type='line', width='1.0')
        
        label = ET.SubElement(generic_node, 'y:NodeLabel', 
                             alignment='center', autoSizePolicy='content',
                             fontFamily='Dialog', fontSize='13', fontStyle='bold',
                             hasBackgroundColor='false', hasLineColor='false',
                             modelName='custom', textColor='#000000')
        label.text = entity
        
        style_props = ET.SubElement(generic_node, 'y:StyleProperties')
        prop = ET.SubElement(style_props, 'y:Property', 
                            class_='java.lang.Boolean',
                            name='y.view.ShadowNodePainter.SHADOW_PAINTING')
        prop.set('value', 'false')
    
    # ===== 2. Атрибуты (овалы) =====
    for entity, info in schema_info.items():
        entity_idx = list(entities).index(entity)
        row = entity_idx // cols
        col = entity_idx % cols
        entity_x = start_x + col * step_x + 60  # центр сущности
        entity_y = start_y + row * step_y + 25   # центр сущности
        
        num_attrs = len(info['columns'])
        radius = 130  # радиус расположения атрибутов вокруг сущности
        
        for attr_idx, column in enumerate(info['columns']):
            attr_name = column['column_name']
            is_pk = attr_name in info['primary_keys']
            is_fk = any(fk['column_name'] == attr_name for fk in info['foreign_keys'])
            
            # Формируем отображаемое имя
            if is_pk and is_fk:
                display_name = f"{attr_name} (PK,FK)"
            elif is_pk:
                display_name = f"{attr_name} (PK)"
            elif is_fk:
                display_name = f"{attr_name} (FK)"
            else:
                display_name = attr_name
            
            # Вычисляем позицию по кругу
            angle = (2 * math.pi * attr_idx / num_attrs) - math.pi/2
            attr_x = entity_x + radius * math.cos(angle) - 40
            attr_y = entity_y + radius * math.sin(angle) - 20
            
            attr_node_id = f'n{node_counter}'
            all_nodes[f"attr_{entity}_{attr_name}"] = attr_node_id
            node_counter += 1
            
            node = ET.SubElement(graph, 'node', id=attr_node_id)
            data = ET.SubElement(node, 'data', key='d6')
            
            # GenericNode для атрибута (овал)
            generic_node = ET.SubElement(data, 'y:GenericNode',
                                        configuration='com.yworks.entityRelationship.attribute')
            
            geometry = ET.SubElement(generic_node, 'y:Geometry',
                                    height='35.0', width='100.0',
                                    x=str(attr_x), y=str(attr_y))
            
            fill = ET.SubElement(generic_node, 'y:Fill',
                                color='#FFFFFF', color2='#FFFFFF', transparent='false')
            
            border = ET.SubElement(generic_node, 'y:BorderStyle',
                                  color='#000000', type='line', width='1.0')
            
            label = ET.SubElement(generic_node, 'y:NodeLabel',
                                 alignment='center', autoSizePolicy='content',
                                 fontFamily='Dialog', fontSize='10', fontStyle='plain',
                                 hasBackgroundColor='false', hasLineColor='false',
                                 modelName='custom', textColor='#000000')
            label.text = display_name
            label.set('x', '5.0')
            label.set('y', '10.0')
            
            style_props = ET.SubElement(generic_node, 'y:StyleProperties')
            prop = ET.SubElement(style_props, 'y:Property',
                                class_='java.lang.Boolean',
                                name='y.view.ShadowNodePainter.SHADOW_PAINTING')
            prop.set('value', 'false')
    
    # ===== 3. Связи между сущностью и атрибутами =====
    edge_counter = 0
    
    for entity, info in schema_info.items():
        entity_node_id = all_nodes[f"entity_{entity}"]
        
        for column in info['columns']:
            attr_name = column['column_name']
            attr_node_id = all_nodes.get(f"attr_{entity}_{attr_name}")
            
            if attr_node_id:
                edge_id = f'e{edge_counter}'
                edge_counter += 1
                
                edge = ET.SubElement(graph, 'edge', id=edge_id, 
                                    source=entity_node_id, target=attr_node_id)
                data = ET.SubElement(edge, 'data', key='d10')
                
                poly_edge = ET.SubElement(data, 'y:PolyLineEdge')
                path = ET.SubElement(poly_edge, 'y:Path', sx='0.0', sy='0.0', tx='0.0', ty='0.0')
                line_style = ET.SubElement(poly_edge, 'y:LineStyle', 
                                          color='#000000', type='line', width='1.0')
                arrows = ET.SubElement(poly_edge, 'y:Arrows', source='none', target='none')
                bend_style = ET.SubElement(poly_edge, 'y:BendStyle', smoothed='false')
    
    # ===== 4. Связи между сущностями (foreign keys) =====
    for rel in relationships:
        from_node_id = all_nodes.get(f"entity_{rel['from_entity']}")
        to_node_id = all_nodes.get(f"entity_{rel['to_entity']}")
        
        if from_node_id and to_node_id:
            # Добавляем ромб для связи
            # Находим позиции сущностей
            from_idx = list(entities).index(rel['from_entity'])
            to_idx = list(entities).index(rel['to_entity'])
            from_row, from_col = divmod(from_idx, cols)
            to_row, to_col = divmod(to_idx, cols)
            
            from_x = start_x + from_col * step_x + 60
            from_y = start_y + from_row * step_y + 25
            to_x = start_x + to_col * step_x + 60
            to_y = start_y + to_row * step_y + 25
            
            # Позиция ромба между сущностями
            rel_x = (from_x + to_x) / 2 - 30
            rel_y = (from_y + to_y) / 2 - 20
            
            rel_node_id = f'n{node_counter}'
            node_counter += 1
            
            node = ET.SubElement(graph, 'node', id=rel_node_id)
            data = ET.SubElement(node, 'data', key='d6')
            
            # GenericNode для связи (ромб)
            generic_node = ET.SubElement(data, 'y:GenericNode',
                                        configuration='com.yworks.entityRelationship.relationship')
            
            geometry = ET.SubElement(generic_node, 'y:Geometry',
                                    height='50.0', width='80.0',
                                    x=str(rel_x), y=str(rel_y))
            
            fill = ET.SubElement(generic_node, 'y:Fill',
                                color='#FFFFFF', color2='#FFFFFF', transparent='false')
            
            border = ET.SubElement(generic_node, 'y:BorderStyle',
                                  color='#000000', type='line', width='1.0')
            
            # Выбираем глагол для связи
            verb = "ссылается на"
            if "user" in rel['from_entity'] and "subscription" in rel['to_entity']:
                verb = "оформляет"
            elif "subscription" in rel['from_entity'] and "plan" in rel['to_entity']:
                verb = "использует"
            elif "promocode" in rel['from_entity'] or "promocode" in rel['to_entity']:
                verb = "активирует"
            
            label = ET.SubElement(generic_node, 'y:NodeLabel',
                                 alignment='center', autoSizePolicy='content',
                                 fontFamily='Dialog', fontSize='11', fontStyle='bold',
                                 hasBackgroundColor='false', hasLineColor='false',
                                 modelName='custom', textColor='#000000')
            label.text = verb
            label.set('x', '15.0')
            label.set('y', '15.0')
            
            style_props = ET.SubElement(generic_node, 'y:StyleProperties')
            prop = ET.SubElement(style_props, 'y:Property',
                                class_='java.lang.Boolean',
                                name='y.view.ShadowNodePainter.SHADOW_PAINTING')
            prop.set('value', 'false')
            
            # Связь от первой сущности к ромбу
            edge1_id = f'e{edge_counter}'
            edge_counter += 1
            edge1 = ET.SubElement(graph, 'edge', id=edge1_id, 
                                 source=from_node_id, target=rel_node_id)
            data1 = ET.SubElement(edge1, 'data', key='d10')
            poly_edge1 = ET.SubElement(data1, 'y:PolyLineEdge')
            path1 = ET.SubElement(poly_edge1, 'y:Path', sx='0.0', sy='0.0', tx='0.0', ty='0.0')
            line_style1 = ET.SubElement(poly_edge1, 'y:LineStyle', color='#000000', type='line', width='1.0')
            arrows1 = ET.SubElement(poly_edge1, 'y:Arrows', source='none', target='none')
            
            # Связь от ромба ко второй сущности
            edge2_id = f'e{edge_counter}'
            edge_counter += 1
            edge2 = ET.SubElement(graph, 'edge', id=edge2_id, 
                                 source=rel_node_id, target=to_node_id)
            data2 = ET.SubElement(edge2, 'data', key='d10')
            poly_edge2 = ET.SubElement(data2, 'y:PolyLineEdge')
            path2 = ET.SubElement(poly_edge2, 'y:Path', sx='0.0', sy='0.0', tx='0.0', ty='0.0')
            line_style2 = ET.SubElement(poly_edge2, 'y:LineStyle', color='#000000', type='line', width='1.0')
            arrows2 = ET.SubElement(poly_edge2, 'y:Arrows', source='none', target='none')
    
    # Сохраняем файл
    tree = ET.ElementTree(graphml)
    ET.indent(tree, '  ')
    tree.write(output_file, encoding='utf-8', xml_declaration=True)
    
    print(f"✅ GraphML файл сохранён: {output_file}")

# ========== АЛЬТЕРНАТИВНЫЙ ПРОСТОЙ ВАРИАНТ ==========

def generate_simple_plantuml(schema_info, relationships, output_file='er_diagram.puml'):
    """Генерирует простой PlantUML файл (работает всегда)"""
    
    puml = []
    puml.append("@startuml")
    puml.append("!theme plain")
    puml.append("hide circle")
    puml.append("")
    
    # Добавляем сущности и их атрибуты
    for entity, info in schema_info.items():
        puml.append(f'entity "{entity}" as {entity.replace(" ", "_")} {{')
        puml.append(f'  * id : <<PK>>')
        
        for column in info['columns']:
            col_name = column['column_name']
            if col_name not in info['primary_keys']:
                puml.append(f'  --')
                puml.append(f'  {col_name} : {column["data_type"]}')
        
        puml.append("}")
        puml.append("")
    
    # Добавляем связи
    for rel in relationships:
        from_entity = rel['from_entity'].replace(" ", "_")
        to_entity = rel['to_entity'].replace(" ", "_")
        puml.append(f'{to_entity} ||--o{{ {from_entity} : "содержит"')
        puml.append("")
    
    puml.append("@enduml")
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write('\n'.join(puml))
    
    print(f"✅ PlantUML файл сохранён: {output_file} (альтернативный вариант)")

# ========== ОСНОВНОЙ БЛОК ==========

def main():
    print("=" * 60)
    print("Генератор ER-диаграмм для yEd")
    print("=" * 60)
    
    print("\n🔄 Подключение к базе данных...")
    
    try:
        # Получаем схему БД
        schema_info = get_db_schema()
        print(f"✅ Найдено таблиц: {len(schema_info)}")
        
        if len(schema_info) == 0:
            print("❌ Таблицы не найдены. Проверьте подключение к БД.")
            return
        
        # Получаем связи
        relationships = get_relationships(schema_info)
        print(f"✅ Найдено связей: {len(relationships)}")
        
        # Генерируем GraphML для yEd
        print("\n📊 Генерация GraphML для yEd...")
        generate_valid_graphml(schema_info, relationships)
        
        # Также генерируем PlantUML как запасной вариант
        generate_simple_plantuml(schema_info, relationships)
        
        print("\n" + "=" * 60)
        print("✨ Готово!")
        print("=" * 60)
        print("\n📁 Созданы файлы:")
        print("   - er_diagram.graphml (открыть в yEd)")
        print("   - er_diagram.puml (открыть в PlantUML)")
        print("\n📖 Инструкция для yEd:")
        print("   1. Откройте yEd Graph Editor")
        print("   2. Файл → Импорт → GraphML...")
        print("   3. Выберите er_diagram.graphml")
        print("   4. Используйте Layout → Hierarchical для авто-размещения")
        
    except psycopg2.Error as e:
        print(f"\n❌ Ошибка подключения к БД: {e}")
        print("\n🔧 Проверьте параметры подключения в DB_CONFIG")
    except Exception as e:
        print(f"\n❌ Ошибка: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()