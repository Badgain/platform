package pool

import (
	"fmt"
	"platform/database/models"
)

const (
	DB_POOL_MAX_CAPACITY = 64
)

var (
	GlobalDbConnectionPool *DbConnectionsPool
)

type DbConnectionsPool struct {
	pool      []*DbConnectionPoolUnit
	size      int
	capacity  int
	freeCount int
}

type DbConnectionPoolUnit struct {
	unit   models.DbConnection
	status bool
}

func (p *DbConnectionsPool) Add(dbc models.DbConnection) {

	unit := &DbConnectionPoolUnit{
		status: true,
		unit:   dbc,
	}
	p.pool = append(p.pool, unit)
	p.freeCount++
}

func (p *DbConnectionsPool) Init() {
	p.size = 0
	p.capacity = DB_POOL_MAX_CAPACITY
	p.pool = make([]*DbConnectionPoolUnit, 0)
	p.freeCount = 0
}

func (p *DbConnectionsPool) GetFree() models.DbConnection {
	for _, unit := range p.pool {
		if unit.status {
			unit.status = false
			p.freeCount--
			return unit.unit
		}
	}
	return nil
}

func (p *DbConnectionsPool) Free(dbc models.DbConnection) int {
	for idx, unit := range p.pool {
		if unit.unit == dbc {
			unit.status = true
			p.freeCount++
			return idx
		}
	}
	return -1
}

func (p *DbConnectionsPool) FreeSize() int {
	fmt.Println(p.freeCount)
	return p.freeCount
}

func (p *DbConnectionsPool) Shutdown() {
	p.size = 0
	p.capacity = 0
	p.freeCount = 0
	p.pool = make([]*DbConnectionPoolUnit, 0)
}

func InitConnectionPool(inst models.DbConnection, pool *DbConnectionsPool, InitConnChan chan<- error) {
	fmt.Println("Init DCP...")
	pool.Init()
	for i := 0; i < pool.capacity; i++ {
		cp, err := inst.GetCopy()
		if err != nil {
			InitConnChan <- err
			pool.Shutdown()
			return
		}
		pool.Add(cp)
	}
	InitConnChan <- nil
}
