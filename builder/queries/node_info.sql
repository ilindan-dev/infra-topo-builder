-- name: UpsertNodesInfo :batchexec
INSERT INTO nodes_info (node_id, serial_number, part_number, revision, product_name,
                        endianness, enable_endianness_per_job, reproducibility_disable)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (node_id) DO UPDATE SET serial_number             = COALESCE(EXCLUDED.serial_number, nodes_info.serial_number),
                                    part_number               = COALESCE(EXCLUDED.part_number, nodes_info.part_number),
                                    revision                  = COALESCE(EXCLUDED.revision, nodes_info.revision),
                                    product_name              = COALESCE(EXCLUDED.product_name, nodes_info.product_name),
                                    endianness                = COALESCE(EXCLUDED.endianness, nodes_info.endianness),
                                    enable_endianness_per_job = COALESCE(EXCLUDED.enable_endianness_per_job,
                                                                         nodes_info.enable_endianness_per_job),
                                    reproducibility_disable   = COALESCE(EXCLUDED.reproducibility_disable,
                                                                         nodes_info.reproducibility_disable);
-- name: GetPortsByNodeID :many
SELECT *
FROM ports
WHERE node_id = $1
ORDER BY port_num;