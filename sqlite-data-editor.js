/**
 * SQLite JSON Data Editor
 * JSON 기반 SQLite 데이터의 CRUD 작업을 위한 편집기
 */
class SQLiteDataEditor {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.currentTable = null;
        this.tableSchemas = new Map(); // 테이블별 스키마 저장
        this.data = new Map(); // 임시 데이터 저장소 (실제 DB 대신)
        
        this.init();
    }

    init() {
        this.render();
        this.attachEventListeners();
    }

    render() {
        this.container.innerHTML = `
            <div class="h-full flex flex-col">
                <!-- Header -->
                <div class="flex items-center justify-between p-4 border-b bg-gray-50">
                    <div>
                        <h3 class="text-lg font-semibold">SQLite 데이터 편집기</h3>
                        <p class="text-sm text-gray-600">JSON 기반 데이터 관리</p>
                    </div>
                    <div class="flex gap-2">
                        <select id="tableSelect" class="border rounded px-3 py-2 text-sm">
                            <option value="">테이블 선택...</option>
                        </select>
                        <button id="refreshData" class="bg-blue-600 text-white px-3 py-2 rounded text-sm hover:bg-blue-700">
                            새로고침
                        </button>
                        <button id="addRecord" class="bg-green-600 text-white px-3 py-2 rounded text-sm hover:bg-green-700" disabled>
                            + 데이터 추가
                        </button>
                    </div>
                </div>

                <!-- Content Area -->
                <div class="flex-1 flex">
                    <!-- Data Table -->
                    <div class="flex-1 overflow-auto">
                        <div id="dataTableContainer" class="h-full">
                            <div class="flex items-center justify-center h-full text-gray-500">
                                테이블을 선택하세요
                            </div>
                        </div>
                    </div>

                    <!-- Editor Panel -->
                    <div id="editorPanel" class="w-96 border-l bg-white hidden">
                        <div class="p-4 border-b">
                            <h4 class="font-medium" id="editorTitle">데이터 편집</h4>
                        </div>
                        <div class="p-4 overflow-auto" id="editorForm">
                            <!-- 동적 폼이 여기에 생성됨 -->
                        </div>
                        <div class="p-4 border-t">
                            <button id="fillSampleData" class="w-full bg-orange-600 text-white py-2 px-3 rounded hover:bg-orange-700 text-sm mb-2">
                                🎯 샘플 데이터로 채우기
                            </button>
                            <div class="flex gap-2">
                                <button id="saveRecord" class="flex-1 bg-blue-600 text-white py-2 px-3 rounded hover:bg-blue-700">
                                    저장
                                </button>
                                <button id="cancelEdit" class="flex-1 bg-gray-600 text-white py-2 px-3 rounded hover:bg-gray-700">
                                    취소
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Delete Confirmation Modal -->
            <div id="deleteModal" class="fixed inset-0 bg-black bg-opacity-50 hidden z-50">
                <div class="flex items-center justify-center h-full">
                    <div class="bg-white rounded-lg p-6 w-96">
                        <h3 class="text-lg font-semibold mb-4">삭제 확인</h3>
                        <p class="text-gray-600 mb-6">정말로 이 데이터를 삭제하시겠습니까?</p>
                        <div class="flex gap-2">
                            <button id="confirmDelete" class="flex-1 bg-red-600 text-white py-2 px-4 rounded hover:bg-red-700">삭제</button>
                            <button id="cancelDelete" class="flex-1 bg-gray-600 text-white py-2 px-4 rounded hover:bg-gray-700">취소</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    attachEventListeners() {
        // Table selection
        document.getElementById('tableSelect').addEventListener('change', (e) => {
            this.selectTable(e.target.value);
        });

        // Refresh button
        document.getElementById('refreshData').addEventListener('click', () => {
            this.loadTableData();
        });

        // Add record button
        document.getElementById('addRecord').addEventListener('click', () => {
            this.showAddForm();
        });

        // Editor form buttons
        document.getElementById('saveRecord').addEventListener('click', () => {
            this.saveRecord();
        });

        document.getElementById('cancelEdit').addEventListener('click', () => {
            this.hideEditor();
        });

        // Fill sample data button
        document.getElementById('fillSampleData').addEventListener('click', () => {
            this.fillFormWithSampleData();
        });

        // Delete modal buttons
        document.getElementById('confirmDelete').addEventListener('click', () => {
            this.confirmDelete();
        });

        document.getElementById('cancelDelete').addEventListener('click', () => {
            this.hideDeleteModal();
        });
    }

    // 테이블 목록 업데이트
    updateTableList(tables) {
        const select = document.getElementById('tableSelect');
        select.innerHTML = '<option value="">테이블 선택...</option>';
        
        tables.forEach(table => {
            const option = document.createElement('option');
            option.value = table.name;
            option.textContent = table.name;
            select.appendChild(option);
        });
    }

    // 테이블 스키마 설정
    setTableSchema(tableName, schema) {
        if (!this.data.has(tableName)) {
            this.data.set(tableName, []);
        }
        // 각 테이블별로 독립적인 스키마 저장
        this.tableSchemas.set(tableName, { name: tableName, ...schema });
    }

    // 테이블 선택
    selectTable(tableName) {
        if (!tableName) {
            this.currentTable = null;
            this.showEmptyState();
            document.getElementById('addRecord').disabled = true;
            return;
        }

        this.currentTable = tableName;
        document.getElementById('addRecord').disabled = false;
        
        // 선택된 테이블이 스키마에 없으면 에러 표시
        if (!this.tableSchemas.has(tableName)) {
            this.showError(`테이블 "${tableName}"의 스키마를 찾을 수 없습니다.`);
            return;
        }
        
        this.loadTableData();
    }

    // 현재 테이블 스키마 가져오기
    getCurrentTableSchema() {
        return this.currentTable ? this.tableSchemas.get(this.currentTable) : null;
    }

    // 테이블 데이터 로드
    loadTableData() {
        if (!this.currentTable) return;

        const tableData = this.data.get(this.currentTable) || [];
        this.renderDataTable(tableData);
    }

    // 데이터 테이블 렌더링
    renderDataTable(data) {
        const container = document.getElementById('dataTableContainer');
        
        if (data.length === 0) {
            container.innerHTML = `
                <div class="flex items-center justify-center h-full text-gray-500">
                    <div class="text-center">
                        <p>데이터가 없습니다</p>
                        <button class="mt-2 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700" onclick="window.dataEditor.showAddForm()">
                            첫 번째 데이터 추가
                        </button>
                    </div>
                </div>
            `;
            return;
        }

        // 테이블 헤더 생성
        const sampleData = data[0].data;
        const columns = Object.keys(sampleData);

        let html = `
            <div class="overflow-auto h-full">
                <table class="w-full text-sm">
                    <thead class="bg-gray-100 sticky top-0">
                        <tr>
                            <th class="px-4 py-2 text-left">ID</th>
                            ${columns.map(col => `<th class="px-4 py-2 text-left">${col}</th>`).join('')}
                            <th class="px-4 py-2 text-left">생성일</th>
                            <th class="px-4 py-2 text-left">수정일</th>
                            <th class="px-4 py-2 text-center">작업</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        data.forEach(record => {
            html += `
                <tr class="border-b hover:bg-gray-50">
                    <td class="px-4 py-2">${record.id}</td>
                    ${columns.map(col => {
                        const value = record.data[col];
                        const displayValue = typeof value === 'object' 
                            ? JSON.stringify(value).substring(0, 50) + '...'
                            : String(value).substring(0, 50);
                        return `<td class="px-4 py-2" title="${JSON.stringify(value)}">${displayValue}</td>`;
                    }).join('')}
                    <td class="px-4 py-2 text-gray-500">${record.created_at}</td>
                    <td class="px-4 py-2 text-gray-500">${record.updated_at}</td>
                    <td class="px-4 py-2">
                        <div class="flex gap-1 justify-center">
                            <button onclick="window.dataEditor.editRecord(${record.id})" 
                                    class="text-blue-600 hover:bg-blue-100 px-2 py-1 rounded text-xs">편집</button>
                            <button onclick="window.dataEditor.deleteRecord(${record.id})" 
                                    class="text-red-600 hover:bg-red-100 px-2 py-1 rounded text-xs">삭제</button>
                        </div>
                    </td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;

        container.innerHTML = html;
    }

    // 빈 상태 표시
    showEmptyState() {
        document.getElementById('dataTableContainer').innerHTML = `
            <div class="flex items-center justify-center h-full text-gray-500">
                테이블을 선택하세요
            </div>
        `;
    }

    // 추가 폼 표시
    showAddForm() {
        this.currentEditId = null;
        const zeroValueData = this.generateZeroValueData();
        this.showEditor('새 데이터 추가', this.generateJSONForm(zeroValueData));
    }

    // 편집 폼 표시
    editRecord(id) {
        const tableData = this.data.get(this.currentTable);
        const record = tableData.find(r => r.id === id);
        if (!record) return;

        this.currentEditId = id;
        this.showEditor('데이터 편집', this.generateJSONForm(record.data));
    }

    // 에디터 표시
    showEditor(title, formHTML) {
        document.getElementById('editorTitle').textContent = title;
        document.getElementById('editorForm').innerHTML = formHTML;
        document.getElementById('editorPanel').classList.remove('hidden');
        
        // JSON 입력 필드에 실시간 유효성 검사 추가
        const jsonInput = document.getElementById('jsonDataInput');
        if (jsonInput) {
            jsonInput.addEventListener('input', () => {
                this.validateJSON();
            });
            
            // 초기 유효성 검사
            setTimeout(() => this.validateJSON(), 100);
        }
    }

    // 에디터 숨기기
    hideEditor() {
        document.getElementById('editorPanel').classList.add('hidden');
        this.currentEditId = null;
    }

    // JSON 폼 생성
    generateJSONForm(data = {}) {
        const currentSchema = this.getCurrentTableSchema();
        if (!currentSchema || !currentSchema.fields) {
            return '<p class="text-gray-500">스키마 정보를 찾을 수 없습니다.</p>';
        }

        const jsonString = JSON.stringify(data, null, 2);
        
        return `
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium mb-2">
                        JSON 데이터
                        <span class="text-gray-500 text-xs">(JSON 형식으로 편집하세요)</span>
                    </label>
                    <textarea id="jsonDataInput" rows="12" 
                              class="w-full border rounded px-3 py-2 font-mono text-sm"
                              placeholder="JSON 형태로 데이터를 입력하세요">${jsonString}</textarea>
                </div>
                
                <!-- 스키마 참조 -->
                <div class="text-xs text-gray-600 bg-gray-50 p-3 rounded">
                    <div class="font-medium mb-2">📋 필드 스키마:</div>
                    <div class="space-y-1">
                        ${currentSchema.fields.map(field => `
                            <div class="flex items-center">
                                <span class="inline-block w-3 h-3 rounded-full mr-2 ${this.getFieldColorClass(field.type)}"></span>
                                <code class="font-mono text-xs">${field.name}</code>
                                <span class="text-gray-500 ml-1">(${field.type})</span>
                                ${field.required ? '<span class="text-red-500 ml-1">*</span>' : ''}
                            </div>
                        `).join('')}
                    </div>
                </div>

                <!-- 유효성 검사 결과 -->
                <div id="validationResult" class="hidden">
                    <div class="text-sm p-3 rounded border"></div>
                </div>
            </div>
        `;
    }

    // 필드 색상 클래스 반환
    getFieldColorClass(fieldType) {
        const colors = {
            string: 'bg-green-500',
            integer: 'bg-yellow-500',
            boolean: 'bg-purple-500',
            datetime: 'bg-red-500',
            object: 'bg-indigo-500',
            array: 'bg-orange-500'
        };
        return colors[fieldType] || 'bg-gray-500';
    }

    // Zero value 데이터 생성
    generateZeroValueData() {
        const currentSchema = this.getCurrentTableSchema();
        if (!currentSchema || !currentSchema.fields) {
            return {};
        }

        const zeroData = {};
        currentSchema.fields.forEach(field => {
            switch (field.type) {
                case 'string':
                    zeroData[field.name] = '';
                    break;
                case 'integer':
                    zeroData[field.name] = 0;
                    break;
                case 'boolean':
                    zeroData[field.name] = false;
                    break;
                case 'datetime':
                    zeroData[field.name] = new Date().toISOString();
                    break;
                case 'object':
                    zeroData[field.name] = {};
                    break;
                case 'array':
                    zeroData[field.name] = [];
                    break;
                default:
                    zeroData[field.name] = null;
            }
        });

        return zeroData;
    }

    // 샘플 데이터 생성
    generateSampleFormData() {
        const currentSchema = this.getCurrentTableSchema();
        if (!currentSchema || !currentSchema.fields) {
            return {};
        }

        const sampleData = {};
        currentSchema.fields.forEach(field => {
            switch (field.type) {
                case 'string':
                    sampleData[field.name] = `sample_${field.name}`;
                    break;
                case 'integer':
                    sampleData[field.name] = Math.floor(Math.random() * 100) + 1;
                    break;
                case 'boolean':
                    sampleData[field.name] = Math.random() > 0.5;
                    break;
                case 'datetime':
                    sampleData[field.name] = new Date().toISOString();
                    break;
                case 'object':
                    if (field.fields && field.fields.length > 0) {
                        sampleData[field.name] = this.generateNestedSampleData(field.fields);
                    } else {
                        sampleData[field.name] = { key: 'value', example: true };
                    }
                    break;
                case 'array':
                    if (field.itemType === 'object' && field.fields && field.fields.length > 0) {
                        sampleData[field.name] = [this.generateNestedSampleData(field.fields)];
                    } else {
                        sampleData[field.name] = field.itemType === 'string' ? ['item1', 'item2'] 
                                               : field.itemType === 'integer' ? [1, 2, 3]
                                               : field.itemType === 'boolean' ? [true, false]
                                               : ['sample'];
                    }
                    break;
                default:
                    sampleData[field.name] = null;
            }
        });

        return sampleData;
    }

    // 중첩된 샘플 데이터 생성
    generateNestedSampleData(fields) {
        const nestedData = {};
        fields.forEach(field => {
            switch (field.type) {
                case 'string':
                    nestedData[field.name] = `nested_${field.name}`;
                    break;
                case 'integer':
                    nestedData[field.name] = Math.floor(Math.random() * 50) + 1;
                    break;
                case 'boolean':
                    nestedData[field.name] = Math.random() > 0.5;
                    break;
                case 'datetime':
                    nestedData[field.name] = new Date().toISOString();
                    break;
                case 'object':
                    nestedData[field.name] = { nested: true };
                    break;
                case 'array':
                    nestedData[field.name] = ['nested_item'];
                    break;
                default:
                    nestedData[field.name] = null;
            }
        });
        return nestedData;
    }

    // 폼에 샘플 데이터 채우기
    fillFormWithSampleData() {
        const sampleData = this.generateSampleFormData();
        const jsonInput = document.getElementById('jsonDataInput');
        if (jsonInput) {
            jsonInput.value = JSON.stringify(sampleData, null, 2);
            this.validateJSON();
        }
    }

    // JSON 유효성 검사
    validateJSON() {
        const jsonInput = document.getElementById('jsonDataInput');
        const resultDiv = document.getElementById('validationResult');
        
        if (!jsonInput || !resultDiv) return;

        try {
            const jsonData = JSON.parse(jsonInput.value);
            resultDiv.className = 'block';
            resultDiv.innerHTML = '<div class="text-sm p-3 rounded border bg-green-50 border-green-200 text-green-700">✅ 유효한 JSON 형식입니다.</div>';
            return true;
        } catch (error) {
            resultDiv.className = 'block';
            resultDiv.innerHTML = `<div class="text-sm p-3 rounded border bg-red-50 border-red-200 text-red-700">❌ JSON 형식 오류: ${error.message}</div>`;
            return false;
        }
    }

    // 폼 필드 생성 (이전 버전, 호환성 유지)
    generateFormFields(data = {}) {
        const currentSchema = this.getCurrentTableSchema();
        if (!currentSchema || !currentSchema.fields) {
            return '<p class="text-gray-500">스키마 정보를 찾을 수 없습니다.</p>';
        }

        let html = '<div class="space-y-4">';
        
        currentSchema.fields.forEach(field => {
            const value = data[field.name] || '';
            html += this.generateFieldHTML(field, value);
        });

        html += '</div>';
        return html;
    }

    // 개별 필드 HTML 생성
    generateFieldHTML(field, value) {
        const fieldId = `field_${field.name}`;
        let inputHTML = '';

        switch (field.type) {
            case 'string':
                inputHTML = `<input type="text" id="${fieldId}" value="${value}" 
                           class="w-full border rounded px-3 py-2" ${field.required ? 'required' : ''}>`;
                break;
            case 'integer':
                inputHTML = `<input type="number" id="${fieldId}" value="${value}" 
                           class="w-full border rounded px-3 py-2" ${field.required ? 'required' : ''}>`;
                break;
            case 'boolean':
                inputHTML = `<select id="${fieldId}" class="w-full border rounded px-3 py-2" ${field.required ? 'required' : ''}>
                           <option value="true" ${value === true ? 'selected' : ''}>True</option>
                           <option value="false" ${value === false ? 'selected' : ''}>False</option>
                         </select>`;
                break;
            case 'datetime':
                const dateValue = value ? new Date(value).toISOString().slice(0, 16) : '';
                inputHTML = `<input type="datetime-local" id="${fieldId}" value="${dateValue}" 
                           class="w-full border rounded px-3 py-2" ${field.required ? 'required' : ''}>`;
                break;
            case 'object':
                const objValue = typeof value === 'object' ? JSON.stringify(value, null, 2) : value;
                inputHTML = `<textarea id="${fieldId}" rows="4" class="w-full border rounded px-3 py-2" 
                           placeholder="JSON 형식으로 입력하세요" ${field.required ? 'required' : ''}>${objValue}</textarea>`;
                break;
            case 'array':
                const arrValue = Array.isArray(value) ? JSON.stringify(value, null, 2) : value;
                inputHTML = `<textarea id="${fieldId}" rows="3" class="w-full border rounded px-3 py-2" 
                           placeholder="JSON 배열 형식으로 입력하세요" ${field.required ? 'required' : ''}>${arrValue}</textarea>`;
                break;
        }

        return `
            <div>
                <label class="block text-sm font-medium mb-1">
                    ${field.name} 
                    <span class="text-gray-500">(${field.type})</span>
                    ${field.required ? '<span class="text-red-500">*</span>' : ''}
                </label>
                ${inputHTML}
            </div>
        `;
    }

    // 레코드 저장
    saveRecord() {
        const currentSchema = this.getCurrentTableSchema();
        if (!this.currentTable || !currentSchema) return;

        try {
            const formData = this.collectFormData();
            const timestamp = new Date().toISOString();

            let tableData = this.data.get(this.currentTable);
            
            if (this.currentEditId) {
                // 수정
                const index = tableData.findIndex(r => r.id === this.currentEditId);
                if (index !== -1) {
                    tableData[index] = {
                        ...tableData[index],
                        data: formData,
                        updated_at: timestamp
                    };
                }
            } else {
                // 추가
                const newId = tableData.length > 0 ? Math.max(...tableData.map(r => r.id)) + 1 : 1;
                tableData.push({
                    id: newId,
                    data: formData,
                    created_at: timestamp,
                    updated_at: timestamp
                });
            }

            this.data.set(this.currentTable, tableData);
            this.loadTableData();
            this.hideEditor();

            // 성공 메시지
            this.showMessage('데이터가 성공적으로 저장되었습니다.', 'success');
        } catch (error) {
            this.showMessage('데이터 저장 중 오류가 발생했습니다: ' + error.message, 'error');
        }
    }

    // 폼 데이터 수집 (JSON 기반)
    collectFormData() {
        const jsonInput = document.getElementById('jsonDataInput');
        if (!jsonInput) {
            throw new Error('JSON 입력 필드를 찾을 수 없습니다.');
        }

        const jsonString = jsonInput.value.trim();
        if (!jsonString) {
            throw new Error('데이터를 입력해주세요.');
        }

        try {
            const data = JSON.parse(jsonString);
            
            // 스키마 기반 유효성 검사
            this.validateDataAgainstSchema(data);
            
            return data;
        } catch (error) {
            throw new Error(`JSON 파싱 오류: ${error.message}`);
        }
    }

    // 스키마 기반 데이터 유효성 검사
    validateDataAgainstSchema(data) {
        const currentSchema = this.getCurrentTableSchema();
        if (!currentSchema || !currentSchema.fields) {
            return; // 스키마가 없으면 검사 건너뛰기
        }

        const errors = [];

        currentSchema.fields.forEach(field => {
            const value = data[field.name];

            // 필수 필드 검사
            if (field.required && (value === undefined || value === null || value === '')) {
                errors.push(`필수 필드 '${field.name}'이 누락되었습니다.`);
                return;
            }

            // 타입 검사
            if (value !== undefined && value !== null) {
                const isValidType = this.validateFieldType(value, field.type);
                if (!isValidType) {
                    errors.push(`필드 '${field.name}'의 타입이 올바르지 않습니다. 예상: ${field.type}`);
                }
            }
        });

        if (errors.length > 0) {
            throw new Error(errors.join('\n'));
        }
    }

    // 필드 타입 유효성 검사
    validateFieldType(value, expectedType) {
        switch (expectedType) {
            case 'string':
                return typeof value === 'string';
            case 'integer':
                return typeof value === 'number' && Number.isInteger(value);
            case 'boolean':
                return typeof value === 'boolean';
            case 'datetime':
                return typeof value === 'string' && !isNaN(Date.parse(value));
            case 'object':
                return typeof value === 'object' && value !== null && !Array.isArray(value);
            case 'array':
                return Array.isArray(value);
            default:
                return true;
        }
    }

    // 레코드 삭제
    deleteRecord(id) {
        this.deleteId = id;
        document.getElementById('deleteModal').classList.remove('hidden');
    }

    // 삭제 확인
    confirmDelete() {
        if (!this.deleteId) return;

        let tableData = this.data.get(this.currentTable);
        tableData = tableData.filter(r => r.id !== this.deleteId);
        this.data.set(this.currentTable, tableData);

        this.loadTableData();
        this.hideDeleteModal();
        this.showMessage('데이터가 삭제되었습니다.', 'success');
    }

    // 삭제 모달 숨기기
    hideDeleteModal() {
        document.getElementById('deleteModal').classList.add('hidden');
        this.deleteId = null;
    }

    // 메시지 표시
    showMessage(message, type = 'info') {
        // 간단한 토스트 메시지
        const toast = document.createElement('div');
        toast.className = `fixed top-4 right-4 px-6 py-3 rounded shadow-lg text-white z-50 ${
            type === 'success' ? 'bg-green-600' : 
            type === 'error' ? 'bg-red-600' : 'bg-blue-600'
        }`;
        toast.textContent = message;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 3000);
    }

    // 샘플 데이터 생성
    generateSampleData() {
        const currentSchema = this.getCurrentTableSchema();
        if (!this.currentTable || !currentSchema) return;

        const sampleData = {};
        currentSchema.fields.forEach(field => {
            switch (field.type) {
                case 'string':
                    sampleData[field.name] = `Sample ${field.name}`;
                    break;
                case 'integer':
                    sampleData[field.name] = Math.floor(Math.random() * 100) + 1;
                    break;
                case 'boolean':
                    sampleData[field.name] = Math.random() > 0.5;
                    break;
                case 'datetime':
                    sampleData[field.name] = new Date().toISOString();
                    break;
                case 'object':
                    sampleData[field.name] = { key: 'value' };
                    break;
                case 'array':
                    sampleData[field.name] = ['item1', 'item2'];
                    break;
            }
        });

        const timestamp = new Date().toISOString();
        const tableData = this.data.get(this.currentTable);
        const newId = tableData.length > 0 ? Math.max(...tableData.map(r => r.id)) + 1 : 1;

        tableData.push({
            id: newId,
            data: sampleData,
            created_at: timestamp,
            updated_at: timestamp
        });

        this.data.set(this.currentTable, tableData);
        this.loadTableData();
        this.showMessage('샘플 데이터가 추가되었습니다.', 'success');
    }

    // 에러 표시
    showError(message) {
        const container = document.getElementById('dataTableContainer');
        container.innerHTML = `
            <div class="flex items-center justify-center h-full text-red-500">
                <div class="text-center">
                    <p class="text-lg">⚠️ 오류</p>
                    <p class="text-sm mt-2">${message}</p>
                </div>
            </div>
        `;
    }
}

// Export for use in main application
window.SQLiteDataEditor = SQLiteDataEditor;