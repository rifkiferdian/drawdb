package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

//go:embed web/templates/*.html web/static
var webFS embed.FS

type App struct{ db *sql.DB }
type Project struct {
	ID                   uint64
	Name                 string
	Diagram              json.RawMessage
	CreatedAt, UpdatedAt time.Time
}

var emptyDiagram = json.RawMessage(`{"version":1,"engine":"mysql","tables":[],"relations":[],"viewport":{"x":0,"y":0,"zoom":1}}`)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required (copy .env.example to .env)")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err = db.Ping(); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err = migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	tmpl := template.Must(template.ParseFS(webFS, "web/templates/*.html"))
	static, _ := fs.Sub(webFS, "web/static")
	r := gin.Default()
	r.SetHTMLTemplate(tmpl)
	r.StaticFS("/static", http.FS(static))
	a := &App{db: db}
	r.GET("/", a.index)
	r.GET("/projects/:id", a.editor)
	api := r.Group("/api/projects")
	api.POST("", a.create)
	api.PUT("/:id", a.update)
	api.DELETE("/:id", a.remove)
	api.POST("/:id/duplicate", a.duplicate)
	api.PUT("/:id/diagram", a.saveDiagram)
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("DrawDB running at http://localhost%s", addr)
	log.Fatal(r.Run(addr))
}

func migrate(db *sql.DB) error {
	b, err := fs.ReadFile(webFS, "web/templates/base.html")
	_ = b
	if err != nil {
		return err
	}
	q := `CREATE TABLE IF NOT EXISTS projects (id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,name VARCHAR(160) NOT NULL,diagram_json JSON NOT NULL,created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,PRIMARY KEY (id),INDEX idx_projects_updated_at (updated_at)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	_, err = db.Exec(q)
	return err
}
func idParam(c *gin.Context) (uint64, bool) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.JSON(400, gin.H{"error": "invalid project id"})
		return 0, false
	}
	return id, true
}
func (a *App) get(id uint64) (Project, error) {
	var p Project
	err := a.db.QueryRow(`SELECT id,name,diagram_json,created_at,updated_at FROM projects WHERE id=?`, id).Scan(&p.ID, &p.Name, &p.Diagram, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (a *App) projectExists(id uint64) bool {
	var exists bool
	if err := a.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM projects WHERE id=?)`, id).Scan(&exists); err != nil {
		return false
	}
	return exists
}
func (a *App) index(c *gin.Context) {
	rows, e := a.db.Query(`SELECT id,name,diagram_json,created_at,updated_at FROM projects ORDER BY updated_at DESC`)
	if e != nil {
		c.String(500, e.Error())
		return
	}
	defer rows.Close()
	ps := []Project{}
	for rows.Next() {
		var p Project
		if rows.Scan(&p.ID, &p.Name, &p.Diagram, &p.CreatedAt, &p.UpdatedAt) == nil {
			ps = append(ps, p)
		}
	}
	c.HTML(200, "index.html", gin.H{"Title": "Projects", "Projects": ps})
}
func (a *App) editor(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, e := a.get(id)
	if errors.Is(e, sql.ErrNoRows) {
		c.Redirect(302, "/")
		return
	}
	if e != nil {
		c.String(500, e.Error())
		return
	}
	diagram, _ := json.Marshal(string(p.Diagram))
	c.HTML(200, "editor.html", gin.H{"Title": p.Name, "Project": p, "DiagramJS": template.JS(diagram)})
}
func decodeName(c *gin.Context) (string, bool) {
	var in struct {
		Name string `json:"name"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" {
		c.JSON(400, gin.H{"error": "project name is required"})
		return "", false
	}
	return strings.TrimSpace(in.Name), true
}
func (a *App) create(c *gin.Context) {
	n, ok := decodeName(c)
	if !ok {
		return
	}
	r, e := a.db.Exec(`INSERT INTO projects(name,diagram_json) VALUES(?,?)`, n, emptyDiagram)
	if e != nil {
		c.JSON(500, gin.H{"error": e.Error()})
		return
	}
	id, _ := r.LastInsertId()
	c.JSON(201, gin.H{"id": id, "name": n})
}
func (a *App) update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	n, ok := decodeName(c)
	if !ok {
		return
	}
	_, e := a.db.Exec(`UPDATE projects SET name=? WHERE id=?`, n, id)
	if e != nil {
		c.JSON(500, gin.H{"error": e.Error()})
		return
	}
	if !a.projectExists(id) {
		c.JSON(404, gin.H{"error": "project not found"})
		return
	}
	c.JSON(200, gin.H{"id": id, "name": n})
}
func (a *App) remove(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	a.db.Exec(`DELETE FROM projects WHERE id=?`, id)
	c.Status(204)
}
func (a *App) duplicate(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, e := a.get(id)
	if e != nil {
		c.JSON(404, gin.H{"error": "project not found"})
		return
	}
	r, e := a.db.Exec(`INSERT INTO projects(name,diagram_json) VALUES(?,?)`, p.Name+" (Copy)", p.Diagram)
	if e != nil {
		c.JSON(500, gin.H{"error": e.Error()})
		return
	}
	nid, _ := r.LastInsertId()
	c.JSON(201, gin.H{"id": nid})
}
func (a *App) saveDiagram(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	raw, e := c.GetRawData()
	if e != nil || !json.Valid(raw) {
		c.JSON(400, gin.H{"error": "invalid diagram JSON"})
		return
	}
	if len(raw) > 8<<20 {
		c.JSON(413, gin.H{"error": "diagram is too large"})
		return
	}
	_, e = a.db.Exec(`UPDATE projects SET diagram_json=? WHERE id=?`, raw, id)
	if e != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("save failed: %v", e)})
		return
	}
	if !a.projectExists(id) {
		c.JSON(404, gin.H{"error": "project not found"})
		return
	}
	c.JSON(200, gin.H{"savedAt": time.Now().Format(time.RFC3339)})
}
