WITH seed_events (id, title, description, starts_at) AS (
    VALUES
        (
            '10000000-0000-0000-0000-000000000001'::uuid,
            'Koncert testowy',
            'Wydarzenie seed do testow rezerwacji i platnosci',
            NOW() + INTERVAL '14 days'
        ),
        (
            '10000000-0000-0000-0000-000000000002'::uuid,
            'Festiwal testowy',
            'Drugie wydarzenie seed do testow listowania',
            NOW() + INTERVAL '21 days'
        ),
        (
            '10000000-0000-0000-0000-000000000003'::uuid,
            'Stand-up demo',
            'Kameralne wydarzenie do sprawdzenia statusow miejsc',
            NOW() + INTERVAL '28 days'
        )
)
INSERT INTO events (id, title, description, starts_at)
SELECT id, title, description, starts_at
FROM seed_events
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    starts_at = EXCLUDED.starts_at;

WITH seat_source AS (
    SELECT
        event_id,
        seat_row,
        seat_number,
        CASE
            WHEN event_id = '10000000-0000-0000-0000-000000000001'::uuid
                AND seat_row = 'A'
                AND seat_number IN (11, 12)
                THEN 'sold'
            WHEN event_id = '10000000-0000-0000-0000-000000000001'::uuid
                AND seat_row = 'B'
                AND seat_number IN (9, 10)
                THEN 'locked'
            WHEN event_id = '10000000-0000-0000-0000-000000000002'::uuid
                AND seat_row = 'C'
                AND seat_number IN (5, 6)
                THEN 'sold'
            ELSE 'available'
        END AS status
    FROM (
        VALUES
            ('10000000-0000-0000-0000-000000000001'::uuid, ARRAY['A', 'B', 'C', 'D']),
            ('10000000-0000-0000-0000-000000000002'::uuid, ARRAY['A', 'B', 'C']),
            ('10000000-0000-0000-0000-000000000003'::uuid, ARRAY['A', 'B'])
    ) AS event_rows(event_id, seat_rows)
    CROSS JOIN LATERAL unnest(event_rows.seat_rows) AS rows(seat_row)
    CROSS JOIN LATERAL generate_series(1, 12) AS numbers(seat_number)
)
INSERT INTO seats (id, event_id, seat_row, seat_number, status)
SELECT
    (
        substr(md5(event_id::text || ':' || seat_row || ':' || seat_number::text), 1, 8) || '-' ||
        substr(md5(event_id::text || ':' || seat_row || ':' || seat_number::text), 9, 4) || '-' ||
        substr(md5(event_id::text || ':' || seat_row || ':' || seat_number::text), 13, 4) || '-' ||
        substr(md5(event_id::text || ':' || seat_row || ':' || seat_number::text), 17, 4) || '-' ||
        substr(md5(event_id::text || ':' || seat_row || ':' || seat_number::text), 21, 12)
    )::uuid,
    event_id,
    seat_row,
    seat_number,
    status
FROM seat_source
ON CONFLICT (event_id, seat_row, seat_number) DO NOTHING;
