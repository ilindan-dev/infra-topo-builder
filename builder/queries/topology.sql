-- name: GetFullTopology :many
SELECT
    n.id, n.log_id, n.node_desc, n.num_ports, n.node_type,
    n.class_version, n.base_version, n.system_image_guid, n.node_guid, n.port_guid,
    ni.serial_number, ni.part_number, ni.revision, ni.product_name,
    ni.endianness, ni.enable_endianness_per_job, ni.reproducibility_disable,
    s.linear_fdb_cap, s.random_fdb_cap, s.mcast_fdb_cap, s.linear_fdb_top,
    s.def_port, s.def_mcast_pri_port, s.def_mcast_not_pri_port, s.life_time_value,
    s.port_state_change, s.optimized_slvl_mapping, s.lids_per_port, s.part_enf_cap,
    s.inb_enf_cap, s.outb_enf_cap, s.filter_raw_inb_cap, s.filter_raw_outb_cap,
    s.enp0, s.mcast_fdb_top
FROM nodes n
         LEFT JOIN nodes_info ni ON n.id = ni.node_id
         LEFT JOIN switches s ON n.id = s.node_id
WHERE n.log_id = $1;