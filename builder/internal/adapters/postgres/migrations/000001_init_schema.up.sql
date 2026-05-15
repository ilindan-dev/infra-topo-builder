CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

------------------------------------------------------------------------------------------------------------------------
-- SECTION: [DICTIONARIES]
-- Доменные справочники и перечисления (ENUMs)
------------------------------------------------------------------------------------------------------------------------

/**
 * @category Dictionary
 * @typedef {ENUM} log_status
 * @description Статус обработки сессии парсинга лога
 */
CREATE TYPE log_status AS ENUM
(
    'pending',
    'in_progress',
    'success',
    'error'
);

/**
 * @category Dictionary
 * @typedef {ENUM} node_kind
 * @description Тип сетевого оборудования (хост 'host' или свитч 'switch')
 */
CREATE TYPE node_kind AS ENUM
(
    'host',
    'switch'
);

------------------------------------------------------------------------------------------------------------------------
-- SECTION: [TABLES]
-- Основные сущности системы
------------------------------------------------------------------------------------------------------------------------

/**
 * @category TABLES
 * @typedef {TABLE} logs
 * @description Журнал сессий парсинга файлов топологии ibdiagnet. Является мастер-записью для загрузки.
 */
CREATE TABLE logs
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status      log_status               NOT NULL DEFAULT 'pending',
    nodes_count INT                      NOT NULL DEFAULT 0,
    ports_count INT                      NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT
    ON TABLE logs IS 'Журнал сессий парсинга файлов топологии ibdiagnet';
COMMENT
    ON COLUMN logs.status IS 'Текущий статус обработки лога';
COMMENT
    ON COLUMN logs.nodes_count IS 'Агрегированное количество успешно распарсенных узлов';
COMMENT
    ON COLUMN logs.ports_count IS 'Агрегированное количество успешно распарсенных портов';

------------------------------------------------------------------------------------------------------------------------

/**
 * @category TABLES
 * @typedef {TABLE} nodes
 * @description Узлы топологии (коммутаторы и хосты). Связаны с таблицей logs.
 */
CREATE TABLE nodes
(
    id                UUID PRIMARY KEY,
    log_id            UUID      NOT NULL REFERENCES logs (id) ON DELETE CASCADE,
    node_desc         TEXT      NOT NULL,
    num_ports         INT       NOT NULL,
    node_type         node_kind NOT NULL,
    class_version     INT       NOT NULL,
    base_version      INT       NOT NULL,
    system_image_guid TEXT      NOT NULL,
    node_guid         TEXT      NOT NULL,
    port_guid         TEXT      NOT NULL,
    UNIQUE (log_id, node_guid)
);

COMMENT
    ON TABLE nodes IS 'Узлы топологии (коммутаторы и хосты)';
COMMENT
    ON COLUMN nodes.node_desc IS 'Словесное описание узла (например, "SWITCH_1" или "HOST_2")';
COMMENT
    ON COLUMN nodes.num_ports IS 'Количество портов, заявленное на узле';
COMMENT
    ON COLUMN nodes.node_type IS 'Тип сетевого оборудования (host или switch)';
COMMENT
    ON COLUMN nodes.node_guid IS 'Глобальный уникальный идентификатор узла (GUID)';

------------------------------------------------------------------------------------------------------------------------

/**
 * @category TABLES
 * @typedef {TABLE} ports
 * @description Информационная сводка по всем портам InfiniBand, принадлежащим узлам.
 */
CREATE TABLE ports
(
    id                                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id                                  UUID NOT NULL REFERENCES nodes (id) ON DELETE CASCADE,
    node_guid                                TEXT NOT NULL,
    port_guid                                TEXT NOT NULL,
    port_num                                 INT  NOT NULL,
    m_key                                    TEXT NOT NULL,
    gid_prfx                                 TEXT NOT NULL,
    msm_lid                                  INT,
    lid                                      INT,
    cap_msk                                  BIGINT,
    m_key_lease_period                       INT,
    diag_code                                INT,
    link_width_actv                          INT,
    link_width_sup                           INT,
    link_width_en                            INT,
    local_port_num                           INT,
    link_speed_en                            INT,
    link_speed_actv                          INT,
    lmc                                      INT,
    m_key_prot_bits                          INT,
    link_down_def_state                      INT,
    port_phy_state                           INT,
    port_state                               INT,
    link_speed_sup                           INT,
    vl_arb_high_cap                          INT,
    vl_high_limit                            INT,
    init_type                                INT,
    vl_cap                                   INT,
    msmsl                                    INT,
    nmtu                                     INT,
    filter_raw_outb                          INT,
    filter_raw_inb                           INT,
    part_enf_outb                            INT,
    part_enf_inb                             INT,
    op_vls                                   INT,
    hoq_life                                 INT,
    vl_stall_cnt                             INT,
    mtu_cap                                  INT,
    init_type_reply                          INT,
    vl_arb_low_cap                           INT,
    p_key_violations                         INT,
    m_key_violations                         INT,
    subn_tmo                                 INT,
    multicast_p_key_trap_suppression_enabled INT,
    client_reregister                        INT,
    guid_cap                                 INT,
    q_key_violations                         INT,
    max_credit_hint                          INT,
    overrun_errs                             INT,
    local_phy_error                          INT,
    resp_time_value                          TEXT,
    link_round_trip_latency                  TEXT,
    ooo_sl_mask                              TEXT,
    cap_msk2                                 TEXT,
    fec_actv                                 TEXT,
    retrans_actv                             TEXT,
    UNIQUE (node_id, port_num)
);

COMMENT
    ON TABLE ports IS 'Информационная сводка по всем портам InfiniBand, принадлежащим узлам';
COMMENT
    ON COLUMN ports.port_num IS 'Порядковый номер порта на узле';
COMMENT
    ON COLUMN ports.cap_msk IS 'Битовая маска возможностей (Capability Mask)';
COMMENT
    ON COLUMN ports.port_state IS 'Текущее логическое состояние порта (например, Active, Down)';
COMMENT
    ON COLUMN ports.fec_actv IS 'Forward Error Correction (строка, может содержать "N/A")';

------------------------------------------------------------------------------------------------------------------------

/**
 * @category TABLES
 * @typedef {TABLE} nodes_info
 * @description Агрегация дополнительных метаданных об узлах из конфигурационного файла .sharp_an_info.
 */
CREATE TABLE nodes_info
(
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id                   UUID NOT NULL UNIQUE REFERENCES nodes (id) ON DELETE CASCADE,
    serial_number             TEXT,
    part_number               TEXT,
    revision                  TEXT,
    product_name              TEXT,
    endianness                INT,
    enable_endianness_per_job INT,
    reproducibility_disable   INT
);

COMMENT
    ON TABLE nodes_info IS 'Агрегация дополнительных метаданных об узлах из файла .sharp_an_info';
COMMENT
    ON COLUMN nodes_info.serial_number IS 'Серийный номер устройства';
COMMENT
    ON COLUMN nodes_info.product_name IS 'Коммерческое название оборудования (например, Mellanox)';
COMMENT
    ON COLUMN nodes_info.endianness IS 'Порядок следования байтов (обычно 0 - Little-Endian, 1 - Big-Endian)';

-----------------------------------------------------------------------------------------------------------------------

/**
 * @category TABLES
 * @typedef {TABLE} switches
 * @description Таблица для хранения сущностей коммутаторов. Связана с nodes.
 */
CREATE TABLE switches
(
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id                UUID NOT NULL UNIQUE REFERENCES nodes (id) ON DELETE CASCADE,
    node_guid              TEXT NOT NULL,
    linear_fdb_cap         INT,
    random_fdb_cap         INT,
    mcast_fdb_cap          INT,
    linear_fdb_top         INT,
    def_port               INT,
    def_mcast_pri_port     INT,
    def_mcast_not_pri_port INT,
    life_time_value        INT,
    port_state_change      INT,
    optimized_slvl_mapping INT,
    lids_per_port          INT,
    part_enf_cap           INT,
    inb_enf_cap            INT,
    outb_enf_cap           INT,
    filter_raw_inb_cap     INT,
    filter_raw_outb_cap    INT,
    enp0                   INT,
    mcast_fdb_top          INT
);

COMMENT
    ON TABLE switches IS 'Таблица для хранения сущностей коммутаторов. Связана с nodes.';
COMMENT
    ON COLUMN switches.linear_fdb_cap IS 'Лимит Linear FDB, может быть NULL если утилита вернула N/A';