-- 查询 DemoApp 的安装信息
SELECT 
    id,
    product,
    device_id,
    hostname,
    os,
    platform,
    platform_version,
    created_at
FROM installs 
WHERE product = 'DemoApp' 
ORDER BY created_at DESC 
LIMIT 5;

-- 查询所有事件
SELECT 
    id,
    product,
    name,
    timestamp,
    properties,
    device_id,
    session_id,
    created_at
FROM events 
WHERE product = 'DemoApp' 
ORDER BY timestamp DESC;

-- 按事件类型统计
SELECT 
    name, 
    COUNT(*) as count,
    COUNT(DISTINCT device_id) as unique_devices,
    COUNT(DISTINCT session_id) as unique_sessions,
    MIN(timestamp) as first_occurrence,
    MAX(timestamp) as last_occurrence
FROM events 
WHERE product = 'DemoApp'
GROUP BY name
ORDER BY count DESC;

-- 查看设备详细信息
SELECT * 
FROM devices 
WHERE device_id = '67e72c13-e684-5662-ad51-975df7b07f8d';
