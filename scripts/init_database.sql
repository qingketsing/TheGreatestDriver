-- Single Drive 数据库初始化脚本
-- 用于在空数据库中创建所需的所有表和索引

-- ============================================
-- 1. 创建 drivelist 表（文件/目录元数据表）
-- ============================================
CREATE TABLE IF NOT EXISTS drivelist (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,              -- 文件/目录完整路径
    capacity BIGINT NOT NULL,        -- 文件大小（目录为0）
    created_at TIMESTAMPTZ DEFAULT now(),
    
    -- 可选：用于秒传和去重的字段（未来优化）
    -- file_hash TEXT,               -- 文件SHA256哈希值
    -- storage_path TEXT,            -- 物理存储路径
    -- ref_count INT DEFAULT 1,      -- 引用计数
    
    CONSTRAINT drivelist_name_check CHECK (name != '')
);

-- 为 name 字段创建索引（加速查询）
CREATE INDEX IF NOT EXISTS idx_drivelist_name ON drivelist(name);

-- 为 created_at 创建索引（方便按时间查询）
CREATE INDEX IF NOT EXISTS idx_drivelist_created_at ON drivelist(created_at);

COMMENT ON TABLE drivelist IS '文件和目录元数据表';
COMMENT ON COLUMN drivelist.id IS '主键ID';
COMMENT ON COLUMN drivelist.name IS '文件/目录的完整相对路径';
COMMENT ON COLUMN drivelist.capacity IS '文件大小（字节），目录为0';
COMMENT ON COLUMN drivelist.created_at IS '创建时间';

-- ============================================
-- 2. 创建 drivelist_closure 表（闭包表）
-- ============================================
CREATE TABLE IF NOT EXISTS drivelist_closure (
    ancestor INTEGER NOT NULL,       -- 祖先节点ID
    descendant INTEGER NOT NULL,     -- 后代节点ID
    depth INT NOT NULL,              -- 层级深度（0表示自己）
    created_at TIMESTAMPTZ DEFAULT now(),
    
    PRIMARY KEY (ancestor, descendant),
    
    -- 外键约束，级联删除
    FOREIGN KEY (ancestor) REFERENCES drivelist(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant) REFERENCES drivelist(id) ON DELETE CASCADE,
    
    -- 约束检查
    CONSTRAINT closure_depth_check CHECK (depth >= 0)
);

-- 创建索引以优化闭包表查询性能
CREATE INDEX IF NOT EXISTS idx_closure_ancestor ON drivelist_closure(ancestor);
CREATE INDEX IF NOT EXISTS idx_closure_descendant ON drivelist_closure(descendant);
CREATE INDEX IF NOT EXISTS idx_closure_depth ON drivelist_closure(depth);
CREATE INDEX IF NOT EXISTS idx_closure_ancestor_depth ON drivelist_closure(ancestor, depth);

COMMENT ON TABLE drivelist_closure IS '闭包表，用于存储文件树的层级关系';
COMMENT ON COLUMN drivelist_closure.ancestor IS '祖先节点ID（包括自己）';
COMMENT ON COLUMN drivelist_closure.descendant IS '后代节点ID（包括自己）';
COMMENT ON COLUMN drivelist_closure.depth IS '从祖先到后代的层级深度，0表示自己';

-- ============================================
-- 3. 插入测试数据（可选）
-- ============================================
-- 取消下面的注释来插入一些测试数据

/*
-- 插入根目录
INSERT INTO drivelist (name, capacity) VALUES ('test', 0) RETURNING id;
-- 假设返回 id = 1

-- 插入自己到自己的闭包关系
INSERT INTO drivelist_closure (ancestor, descendant, depth) VALUES (1, 1, 0);

-- 插入子文件
INSERT INTO drivelist (name, capacity) VALUES ('test/file1.txt', 1024) RETURNING id;
-- 假设返回 id = 2

INSERT INTO drivelist_closure (ancestor, descendant, depth) VALUES (2, 2, 0);
INSERT INTO drivelist_closure (ancestor, descendant, depth) 
SELECT ancestor, 2, depth + 1 FROM drivelist_closure WHERE descendant = 1;
*/

-- ============================================
-- 4. 查看表结构
-- ============================================
-- \d drivelist
-- \d drivelist_closure

-- ============================================
-- 5. 完成信息
-- ============================================
DO $$
BEGIN
    RAISE NOTICE '✓ 数据库初始化完成！';
    RAISE NOTICE '  - drivelist 表已创建';
    RAISE NOTICE '  - drivelist_closure 表已创建';
    RAISE NOTICE '  - 所有索引已创建';
    RAISE NOTICE '';
    RAISE NOTICE '下一步：启动 Single Drive 服务器';
    RAISE NOTICE '  cd cmd/server && go run main.go';
END $$;
