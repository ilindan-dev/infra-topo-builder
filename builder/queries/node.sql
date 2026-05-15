-- name: InsertNodes :copyfrom
INSERT INTO nodes (id, log_id, node_desc, num_ports, node_type, class_version,
                   base_version, system_image_guid, node_guid, port_guid)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetNodesByLogID :many
SELECT id,
       log_id,
       node_desc,
       num_ports,
       node_type,
       class_version,
       base_version,
       system_image_guid,
       node_guid,
       port_guid
FROM nodes
WHERE log_id = $1;

-- name: GetNodeByID :one
SELECT id,
       log_id,
       node_desc,
       num_ports,
       node_type,
       class_version,
       base_version,
       system_image_guid,
       node_guid,
       port_guid
FROM nodes
WHERE id = $1;
